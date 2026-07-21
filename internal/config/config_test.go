package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLoadAppliesDefaults(t *testing.T) {
	cfg := loadTestConfig(t, `
targets:
  - name: rack-01
    host: https://vertiv.example
    devices:
      - name: ENV_THD
        equip_id: -98
`)

	if cfg.Exporter.ListenAddress != ":9101" {
		t.Fatalf("ListenAddress = %q, want %q", cfg.Exporter.ListenAddress, ":9101")
	}
	if cfg.Exporter.MetricsPath != "/metrics" {
		t.Fatalf("MetricsPath = %q, want %q", cfg.Exporter.MetricsPath, "/metrics")
	}
	if cfg.Exporter.ScrapeTimeout != 10*time.Second {
		t.Fatalf("ScrapeTimeout = %s, want %s", cfg.Exporter.ScrapeTimeout, 10*time.Second)
	}
	if got := cfg.Targets[0].Devices[0].EquipID; got != -98 {
		t.Fatalf("EquipID = %d, want -98", got)
	}
}

func TestLoadReadsExporterSettings(t *testing.T) {
	cfg := loadTestConfig(t, `
exporter:
  listen_address: 127.0.0.1:9200
  metrics_path: /prometheus
  scrape_timeout: 3s
  metrics_file: custom.md
  debug_response: true
targets:
  - name: rack-01
    host: https://vertiv.example
    devices:
      - name: AC1
        type: ac
        equip_id: 23
`)

	if cfg.Exporter.ListenAddress != "127.0.0.1:9200" || cfg.Exporter.MetricsPath != "/prometheus" {
		t.Fatalf("unexpected exporter addresses: %+v", cfg.Exporter)
	}
	if cfg.Exporter.ScrapeTimeout != 3*time.Second || cfg.Exporter.MetricsFile != "custom.md" || !cfg.Exporter.DebugResponse {
		t.Fatalf("unexpected exporter settings: %+v", cfg.Exporter)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "invalid yaml", config: "targets: [", wantErr: "parse config"},
		{name: "no targets", config: "exporter: {}", wantErr: "at least one target"},
		{name: "missing target identity", config: targetConfig("", "https://vertiv.example", "AC1", 23), wantErr: "name and host"},
		{name: "missing target host", config: targetConfig("rack-01", "", "AC1", 23), wantErr: "name and host"},
		{name: "no devices", config: "targets:\n  - name: rack-01\n    host: https://vertiv.example\n", wantErr: "at least one device"},
		{name: "missing device name", config: targetConfig("rack-01", "https://vertiv.example", "", 23), wantErr: "name and equip_id"},
		{name: "zero equip id", config: targetConfig("rack-01", "https://vertiv.example", "AC1", 0), wantErr: "name and equip_id"},
		{name: "relative metrics path", config: exporterConfig("metrics", "10s"), wantErr: "absolute path"},
		{name: "root metrics path", config: exporterConfig("/", "10s"), wantErr: "absolute path"},
		{name: "invalid metrics path", config: exporterConfig("/metrics?format=text", "10s"), wantErr: "invalid characters"},
		{name: "repeated slash in metrics path", config: exporterConfig("//metrics", "10s"), wantErr: "repeated slashes"},
		{name: "dot segment in metrics path", config: exporterConfig("/metrics/.", "10s"), wantErr: "dot segments"},
		{name: "parent segment in metrics path", config: exporterConfig("/metrics/..", "10s"), wantErr: "dot segments"},
		{name: "negative timeout", config: exporterConfig("/metrics", "-1s"), wantErr: "must be positive"},
		{
			name: "duplicate target name",
			config: targetConfig("rack-01", "https://one.example", "AC1", 23) +
				"  - name: rack-01\n    host: https://two.example\n    devices:\n      - name: AC2\n        equip_id: 24\n",
			wantErr: "must be unique",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTestConfig(t, tt.config)
			_, err := Load(path)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAllowsTrailingSlashMetricsPath(t *testing.T) {
	cfg := loadTestConfig(t, exporterConfig("/metrics/", "10s"))
	if got, want := cfg.Exporter.MetricsPath, "/metrics/"; got != want {
		t.Fatalf("MetricsPath = %q, want %q", got, want)
	}
}

func TestLoadReportsReadError(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil || !strings.Contains(err.Error(), "read config") {
		t.Fatalf("Load() error = %v, want read config error", err)
	}
}

func loadTestConfig(t *testing.T, contents string) *Config {
	t.Helper()
	cfg, err := Load(writeTestConfig(t, contents))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return cfg
}

func writeTestConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func targetConfig(targetName, host, deviceName string, equipID int) string {
	return "targets:\n" +
		"  - name: " + targetName + "\n" +
		"    host: " + host + "\n" +
		"    devices:\n" +
		"      - name: " + deviceName + "\n" +
		"        equip_id: " + strconv.Itoa(equipID) + "\n"
}

func exporterConfig(metricsPath, timeout string) string {
	return "exporter:\n" +
		"  metrics_path: " + metricsPath + "\n" +
		"  scrape_timeout: " + timeout + "\n" +
		targetConfig("rack-01", "https://vertiv.example", "AC1", 23)
}
