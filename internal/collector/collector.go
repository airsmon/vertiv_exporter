package collector

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/simplifiedchinese"

	"vertiv_exporter/internal/client"
	"vertiv_exporter/internal/config"
)

var metricLabels = []string{"target", "device", "equip_id"}

type targetState struct {
	target config.Target
	client *client.Client
}

type VertivCollector struct {
	config   *config.Config
	targets  map[string]targetState
	metrics  map[int]MetricDefinition
	descs    map[int]*prometheus.Desc
	acSignal *prometheus.Desc
	thdDescs thdDescs
	upsDescs upsDescs
	upDesc   *prometheus.Desc
	duration prometheus.Histogram
	errors   prometheus.Counter
	ctx      context.Context
	cancel   context.CancelFunc
}

func New(parent context.Context, cfg *config.Config) (*VertivCollector, error) {
	metricDefs, err := LoadMetricDefinitions(cfg.Exporter.MetricsFile)
	if err != nil {
		return nil, err
	}

	targets := make(map[string]targetState, len(cfg.Targets))
	ctx, cancel := context.WithCancel(parent)

	for _, target := range cfg.Targets {
		cl, err := client.NewClient(target.Host, target.TLSSkipVerify, target.Username, target.Password, cfg.Exporter.DebugResponse)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("create client for %q: %w", target.Name, err)
		}
		if err := cl.Login(ctx); err != nil {
			log.Printf("initial login for %s failed: %v", target.Name, err)
		}
		targets[target.Name] = targetState{target: target, client: cl}
	}

	descs := make(map[int]*prometheus.Desc, len(metricDefs))
	for fieldID, def := range metricDefs {
		descs[fieldID] = prometheus.NewDesc(def.Name, def.Help, metricLabels, nil)
	}

	col := &VertivCollector{
		config:   cfg,
		targets:  targets,
		metrics:  metricDefs,
		descs:    descs,
		acSignal: newACSignalDesc(),
		thdDescs: newTHDDescs(),
		upsDescs: newUPSDescs(),
		upDesc: prometheus.NewDesc(
			"vertiv_exporter_up",
			"Whether the target scrape succeeded",
			[]string{"target"},
			nil,
		),
		duration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "vertiv_exporter_scrape_duration_seconds",
			Help:    "Time spent collecting metrics from Vertiv targets",
			Buckets: prometheus.DefBuckets,
		}),
		errors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vertiv_exporter_scrape_failures_total",
			Help: "Total scrape failures while querying Vertiv targets",
		}),
		ctx:    ctx,
		cancel: cancel,
	}

	for _, state := range col.targets {
		go col.startKeepAlive(ctx, state)
	}

	return col, nil
}

func newACSignalDesc() *prometheus.Desc {
	return prometheus.NewDesc(
		"vertiv_ac_signal_value",
		"Raw AC signal value reported by the Vertiv device",
		[]string{"target", "device", "equip_id", "signal_name", "occurrence"},
		nil,
	)
}

func normalizeACSignalName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || utf8.ValidString(raw) {
		return raw
	}

	decoded, err := simplifiedchinese.GB18030.NewDecoder().String(raw)
	if err == nil && utf8.ValidString(decoded) {
		return strings.TrimSpace(decoded)
	}

	return strings.ToValidUTF8(raw, "\uFFFD")
}

func (c *VertivCollector) Close() {
	c.cancel()
}

func (c *VertivCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
	ch <- c.acSignal
	for _, desc := range c.thdDescs.all() {
		ch <- desc
	}
	for _, desc := range c.upsDescs.all() {
		ch <- desc
	}
	ch <- c.upDesc
	c.duration.Describe(ch)
	c.errors.Describe(ch)
}

func (c *VertivCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(c.ctx, c.config.Exporter.ScrapeTimeout)
	defer cancel()

	var (
		mu      sync.Mutex
		emitted []prometheus.Metric
		upMap   = make(map[string]float64, len(c.targets))
	)

	group, ctx := errgroup.WithContext(ctx)
	for _, state := range c.targets {
		state := state
		group.Go(func() error {
			targetUp := 1.0
			for _, device := range state.target.Devices {
				samples, err := state.client.FetchDeviceData(ctx, device.EquipID)
				if err != nil {
					targetUp = 0
					c.errors.Inc()
					log.Printf("scrape failed for %s/%s (equip_id=%d): %v", state.target.Name, device.Name, device.EquipID, err)
					break
				}

				local := c.buildDeviceMetrics(state.target.Name, device, samples)

				mu.Lock()
				emitted = append(emitted, local...)
				mu.Unlock()
			}

			mu.Lock()
			upMap[state.target.Name] = targetUp
			mu.Unlock()
			return nil
		})
	}

	_ = group.Wait()

	for _, metric := range emitted {
		ch <- metric
	}

	for target, up := range upMap {
		ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, up, target)
	}

	c.duration.Observe(time.Since(start).Seconds())
	c.duration.Collect(ch)
	c.errors.Collect(ch)
}

func (c *VertivCollector) buildDeviceMetrics(target string, device config.Device, samples map[int]client.Sample) []prometheus.Metric {
	if isTHDDevice(device) {
		return buildTHDMetrics(c.thdDescs, target, device, samples)
	}
	if isUPSDevice(device) {
		return buildUPSMetrics(c.upsDescs, target, device, samples)
	}

	fieldIDs := make([]int, 0, len(samples))
	for fieldID := range samples {
		fieldIDs = append(fieldIDs, fieldID)
	}
	sort.Ints(fieldIDs)

	metrics := make([]prometheus.Metric, 0, len(samples)*2)
	occurrences := make(map[string]int, len(samples))
	for _, fieldID := range fieldIDs {
		sample := samples[fieldID]
		signalName := normalizeACSignalName(sample.Name)
		if signalName != "" {
			occurrences[signalName]++
			metrics = append(metrics, prometheus.MustNewConstMetric(
				c.acSignal,
				prometheus.GaugeValue,
				sample.Value,
				target,
				device.Name,
				strconv.Itoa(device.EquipID),
				signalName,
				strconv.Itoa(occurrences[signalName]),
			))
		}

		desc, ok := c.descs[fieldID]
		if !ok {
			continue
		}

		metrics = append(metrics, prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			sample.Value,
			target,
			device.Name,
			strconv.Itoa(device.EquipID),
		))
	}

	return metrics
}

func (c *VertivCollector) startKeepAlive(ctx context.Context, state targetState) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := state.client.KeepAlive(ctx); err != nil {
				log.Printf("keepalive failed for %s: %v", state.target.Name, err)
			}
		}
	}
}
