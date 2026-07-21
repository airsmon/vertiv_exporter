package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"vertiv_exporter/internal/config"
)

func TestCollectObservesCurrentScrapeDuration(t *testing.T) {
	duration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "test_scrape_duration_seconds",
		Help: "Test scrape duration",
	})
	col := &VertivCollector{
		config: &config.Config{
			Exporter: config.ExporterConfig{ScrapeTimeout: time.Second},
		},
		targets:  map[string]targetState{},
		duration: duration,
		ctx:      context.Background(),
		errors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_scrape_failures_total",
			Help: "Test scrape failures",
		}),
	}

	metrics := make(chan prometheus.Metric, 2)
	col.Collect(metrics)
	close(metrics)

	for metric := range metrics {
		var value dto.Metric
		if err := metric.Write(&value); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		if value.Histogram == nil {
			continue
		}
		if got, want := value.GetHistogram().GetSampleCount(), uint64(1); got != want {
			t.Fatalf("duration sample count = %d, want %d", got, want)
		}
		return
	}

	t.Fatal("duration histogram was not collected")
}
