package collector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMetricDefinitionsUsesBuiltInDefaults(t *testing.T) {
	defs, err := LoadMetricDefinitions("")
	if err != nil {
		t.Fatalf("LoadMetricDefinitions returned error: %v", err)
	}
	if got, want := len(defs), 181; got != want {
		t.Fatalf("definition count = %d, want %d", got, want)
	}

	seenNames := make(map[string]int, len(defs))
	for fieldID, definition := range defs {
		if definition.FieldID != fieldID {
			t.Fatalf("field %d definition has field id %d", fieldID, definition.FieldID)
		}
		if definition.Name == "" || definition.Help == "" {
			t.Fatalf("field %d has incomplete metadata: %#v", fieldID, definition)
		}
		if previousID, exists := seenNames[definition.Name]; exists {
			t.Fatalf("fields %d and %d use duplicate metric name %q", previousID, fieldID, definition.Name)
		}
		seenNames[definition.Name] = fieldID
	}

	checks := map[int]MetricDefinition{
		2: {
			Name: "vertiv_ac_temperature_return_air_celsius",
			Help: "Return air temperature measurement",
		},
		261: {
			Name: "vertiv_ac_status_compressor_output",
			Help: "Compressor output",
		},
		326: {
			Name: "vertiv_ac_pressure_exhaust_bar",
			Help: "Exhaust pressure measurement",
		},
	}

	for fieldID, want := range checks {
		got, ok := defs[fieldID]
		if !ok {
			t.Fatalf("field %d missing from definitions", fieldID)
		}
		if got.Name != want.Name {
			t.Fatalf("field %d name = %q, want %q", fieldID, got.Name, want.Name)
		}
		if got.Help != want.Help {
			t.Fatalf("field %d help = %q, want %q", fieldID, got.Help, want.Help)
		}
	}
}

func TestLoadMetricDefinitionsReturnsIndependentBuiltInMaps(t *testing.T) {
	first, err := LoadMetricDefinitions("")
	if err != nil {
		t.Fatalf("first LoadMetricDefinitions returned error: %v", err)
	}
	delete(first, 2)

	second, err := LoadMetricDefinitions("")
	if err != nil {
		t.Fatalf("second LoadMetricDefinitions returned error: %v", err)
	}
	if _, ok := second[2]; !ok {
		t.Fatal("modifying one built-in definitions map affected a later load")
	}
}

func TestLoadMetricDefinitionsSupportsMarkdownOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.md")
	content := "| `custom_metric` | 999 | Custom metric help | 42 |\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write metrics override: %v", err)
	}

	defs, err := LoadMetricDefinitions(path)
	if err != nil {
		t.Fatalf("LoadMetricDefinitions returned error: %v", err)
	}
	if got, want := len(defs), 1; got != want {
		t.Fatalf("definition count = %d, want %d", got, want)
	}

	got, ok := defs[999]
	if !ok {
		t.Fatal("custom field 999 missing from definitions")
	}
	if got.Name != "custom_metric" || got.Help != "Custom metric help" {
		t.Fatalf("custom definition = %#v", got)
	}
}
