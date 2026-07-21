package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"vertiv_exporter/internal/collector"
	"vertiv_exporter/internal/config"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("vertiv_exporter", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configFile := flags.String("config.file", "config.yaml", "Path to the YAML configuration file")
	showVersion := flags.Bool("version", false, "Print version information and exit")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintf(stdout, "vertiv_exporter version=%s commit=%s build_date=%s\n", version, commit, buildDate)
		return 0
	}

	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(stderr, "load configuration: %v\n", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	vertivCollector, err := collector.New(ctx, cfg)
	if err != nil {
		fmt.Fprintf(stderr, "create collector: %v\n", err)
		return 1
	}
	defer vertivCollector.Close()

	registry := prometheus.NewRegistry()
	if err := registry.Register(vertivCollector); err != nil {
		fmt.Fprintf(stderr, "register collector: %v\n", err)
		return 1
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Exporter.MetricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<!doctype html><html><body><h1>Vertiv Exporter</h1><p><a href=\"%s\">Metrics</a></p></body></html>\n", html.EscapeString(cfg.Exporter.MetricsPath))
	})

	server := &http.Server{
		Addr:              cfg.Exporter.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	logger := log.New(stderr, "", log.LstdFlags)
	logger.Printf("listening on %s (metrics: %s)", cfg.Exporter.ListenAddress, cfg.Exporter.MetricsPath)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "serve HTTP: %v\n", err)
			return 1
		}
		return 0
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(stderr, "shut down HTTP server: %v\n", err)
		return 1
	}

	return 0
}
