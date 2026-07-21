package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"vertiv_exporter/internal/config"
)

func TestCollectorScrapesCGIIntoPrometheusMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/login.cgi":
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "authenticated", Path: "/"})
			w.WriteHeader(http.StatusOK)

		case "/cgi-bin/p05_equip_sample.cgi":
			if cookie, err := r.Cookie("session"); err != nil || cookie.Value != "authenticated" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(
				"3021,AC_1,ENP_AC_SRVII[COM]^2,Return air temperature measurement,28.6,℃;" +
					"261,Compressor output,OutputMove[1],",
			))

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Exporter: config.ExporterConfig{ScrapeTimeout: 2 * time.Second},
		Targets: []config.Target{
			{
				Name:     "dc-rack-01",
				Host:     server.URL,
				Username: "admin",
				Password: "secret",
				Devices: []config.Device{
					{Name: "AC_1", Type: "ac", EquipID: 23},
				},
			},
		},
	}

	col, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer col.Close()

	registry := prometheus.NewRegistry()
	if err := registry.Register(col); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	assertMetricValue(t, families, "vertiv_ac_temperature_return_air_celsius", 28.6)
	assertMetricValue(t, families, "vertiv_ac_status_compressor_output", 1)
	assertMetricValue(t, families, "vertiv_exporter_up", 1)

	duration := findMetricFamily(families, "vertiv_exporter_scrape_duration_seconds")
	if duration == nil || len(duration.Metric) != 1 {
		t.Fatalf("duration metric family missing or empty")
	}
	if got, want := duration.Metric[0].GetHistogram().GetSampleCount(), uint64(1); got != want {
		t.Fatalf("duration sample count = %d, want %d", got, want)
	}
}

func TestCollectorCloseCancelsInFlightScrape(t *testing.T) {
	requestStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/login.cgi":
			w.WriteHeader(http.StatusOK)
		case "/cgi-bin/p05_equip_sample.cgi":
			close(requestStarted)
			<-r.Context().Done()
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Exporter: config.ExporterConfig{ScrapeTimeout: 30 * time.Second},
		Targets: []config.Target{
			{
				Name: "dc-rack-01",
				Host: server.URL,
				Devices: []config.Device{
					{Name: "AC_1", Type: "ac", EquipID: 23},
				},
			},
		},
	}

	col, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer col.Close()

	registry := prometheus.NewRegistry()
	if err := registry.Register(col); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	gatherDone := make(chan error, 1)
	go func() {
		_, err := registry.Gather()
		gatherDone <- err
	}()

	select {
	case <-requestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("scrape request did not start")
	}

	col.Close()

	select {
	case err := <-gatherDone:
		if err != nil {
			t.Fatalf("gather metrics after Close: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight scrape was not canceled by Close")
	}
}

func assertMetricValue(t *testing.T, families []*dto.MetricFamily, name string, want float64) {
	t.Helper()
	family := findMetricFamily(families, name)
	if family == nil || len(family.Metric) != 1 {
		t.Fatalf("metric family %q missing or empty", name)
	}

	metric := family.Metric[0]
	got := metric.GetGauge().GetValue()
	if got != want {
		t.Fatalf("metric %q value = %v, want %v", name, got, want)
	}
}

func findMetricFamily(families []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}
