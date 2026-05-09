package collector

import "testing"

func TestLoadMetricDefinitionsUsesEmbeddedDefaults(t *testing.T) {
	defs, err := LoadMetricDefinitions("")
	if err != nil {
		t.Fatalf("LoadMetricDefinitions returned error: %v", err)
	}

	checks := map[int]string{
		2:   "vertiv_ac_temperature_return_air_celsius",
		261: "vertiv_ac_status_compressor_output",
		326: "vertiv_ac_pressure_exhaust_bar",
	}

	for fieldID, wantName := range checks {
		got, ok := defs[fieldID]
		if !ok {
			t.Fatalf("field %d missing from definitions", fieldID)
		}
		if got.Name != wantName {
			t.Fatalf("field %d name = %q, want %q", fieldID, got.Name, wantName)
		}
	}
}
