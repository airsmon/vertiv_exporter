package collector

import (
	"testing"

	dto "github.com/prometheus/client_model/go"

	"vertiv_exporter/internal/client"
	"vertiv_exporter/internal/config"
)

func TestBuildTHDMetrics(t *testing.T) {
	samples := map[int]client.Sample{
		3:     {FieldID: 3, Name: "RACK1 Cool Aisle Top Temp", Value: 21.700001},
		5:     {FieldID: 5, Name: "RACK1 Cool Aisle Hum", Value: 40.4},
		12:    {FieldID: 12, Name: "RACK1 Cool Aisle Door Status", Value: 0},
		3010:  {FieldID: 3010, Name: "RACK1 THD Comm Status", Value: 0},
		10000: {FieldID: 10000, Name: "RACK1", Value: 1},
		10012: {FieldID: 10012, Name: "High Temp Alarm Rack Count", Value: 0},
		155:   {FieldID: 155, Name: "RACK PMC Cool Aisle Top Temp", Value: 21.7},
	}

	metrics := buildTHDMetrics(
		newTHDDescs(),
		"dc-rack-01",
		config.Device{Name: "ENV_THD", EquipID: 5005},
		samples,
	)

	if len(metrics) != 7 {
		t.Fatalf("metric count = %d, want 7", len(metrics))
	}
}

func TestBuildDeviceMetricsForTHDProducesExpectedMetric(t *testing.T) {
	col := &VertivCollector{thdDescs: newTHDDescs()}
	device := config.Device{Name: "ENV_THD", EquipID: 5005}
	samples := map[int]client.Sample{
		3: {FieldID: 3, Name: "RACK1 Cool Aisle Top Temp", Value: 23.700001},
		5: {FieldID: 5, Name: "RACK1 Cool Aisle Hum", Value: 40.400002},
	}

	metrics := col.buildDeviceMetrics("dc-rack-01", device, samples)
	if len(metrics) != 2 {
		t.Fatalf("metric count = %d, want 2", len(metrics))
	}

	values := make(map[string]float64, len(metrics))
	for _, metric := range metrics {
		var pb dto.Metric
		if err := metric.Write(&pb); err != nil {
			t.Fatalf("write metric: %v", err)
		}

		key := ""
		for _, label := range pb.GetLabel() {
			if label.GetName() == "__name__" {
				key = label.GetValue()
				break
			}
		}
		if key == "" {
			switch len(pb.GetLabel()) {
			case 6:
				key = "temperature"
			case 5:
				key = "humidity"
			}
		}
		values[key] = pb.GetGauge().GetValue()
	}

	if got := values["temperature"]; got != 23.7 {
		t.Fatalf("temperature metric value = %v, want 23.7", got)
	}
	if got := values["humidity"]; got != 40.4 {
		t.Fatalf("humidity metric value = %v, want 40.4", got)
	}
}
