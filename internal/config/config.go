package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Exporter ExporterConfig `yaml:"exporter"`
	Targets  []Target       `yaml:"targets"`
}

type ExporterConfig struct {
	ListenAddress string        `yaml:"listen_address"`
	MetricsPath   string        `yaml:"metrics_path"`
	ScrapeTimeout time.Duration `yaml:"scrape_timeout"`
	MetricsFile   string        `yaml:"metrics_file"`
	DebugResponse bool          `yaml:"debug_response"`
}

type Target struct {
	Name          string   `yaml:"name"`
	Host          string   `yaml:"host"`
	Username      string   `yaml:"username"`
	Password      string   `yaml:"password"`
	TLSSkipVerify bool     `yaml:"tls_skip_verify"`
	Devices       []Device `yaml:"devices"`
}

type Device struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	EquipID int    `yaml:"equip_id"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if cfg.Exporter.ListenAddress == "" {
		cfg.Exporter.ListenAddress = ":9101"
	}
	if cfg.Exporter.MetricsPath == "" {
		cfg.Exporter.MetricsPath = "/metrics"
	}
	if cfg.Exporter.ScrapeTimeout == 0 {
		cfg.Exporter.ScrapeTimeout = 10 * time.Second
	}
	if len(cfg.Targets) == 0 {
		return nil, fmt.Errorf("config must define at least one target")
	}

	for i, target := range cfg.Targets {
		if target.Name == "" || target.Host == "" {
			return nil, fmt.Errorf("target %d must define name and host", i)
		}
		if len(target.Devices) == 0 {
			return nil, fmt.Errorf("target %q must define at least one device", target.Name)
		}
		for j, device := range target.Devices {
			if device.Name == "" || device.EquipID == 0 {
				return nil, fmt.Errorf("target %q device %d must define name and equip_id", target.Name, j)
			}
		}
	}

	return &cfg, nil
}
