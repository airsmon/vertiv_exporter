package collector

import (
	"testing"

	dto "github.com/prometheus/client_model/go"

	"vertiv_exporter/internal/client"
	"vertiv_exporter/internal/config"
)

func TestBuildUPSMetrics(t *testing.T) {
	samples := map[int]client.Sample{
		2:   {FieldID: 2, Name: "Phase A Input Voltage", Value: 217.800003},
		16:  {FieldID: 16, Name: "Line Ab Input Voltage", Value: 378.399994},
		134: {FieldID: 134, Name: "Local Phase A Output Active Power", Value: 3.11},
		72:  {FieldID: 72, Name: "Battery Current Capacity", Value: 100},
		73:  {FieldID: 73, Name: "Battery Discharge Times", Value: 59},
		25:  {FieldID: 25, Name: "Power Supply", Value: 1},
	}

	metrics := buildUPSMetrics(newUPSDescs(), "dc-rack-01", config.Device{Name: "UPS_1", EquipID: 26}, samples)
	if len(metrics) != 6 {
		t.Fatalf("metric count = %d, want 6", len(metrics))
	}
}

func TestBuildDeviceMetricsForUPSProducesExpectedValues(t *testing.T) {
	col := &VertivCollector{upsDescs: newUPSDescs()}
	device := config.Device{Name: "UPS_1", EquipID: 26}
	samples := map[int]client.Sample{
		2:  {FieldID: 2, Name: "Phase A Input Voltage", Value: 217.800003},
		73: {FieldID: 73, Name: "Battery Discharge Times", Value: 59},
	}

	metrics := col.buildDeviceMetrics("dc-rack-01", device, samples)
	if len(metrics) != 2 {
		t.Fatalf("metric count = %d, want 2", len(metrics))
	}

	var (
		voltageFound   bool
		dischargeFound bool
	)

	for _, metric := range metrics {
		var pb dto.Metric
		if err := metric.Write(&pb); err != nil {
			t.Fatalf("write metric: %v", err)
		}

		switch {
		case pb.Gauge != nil:
			voltageFound = true
			if got := pb.GetGauge().GetValue(); got != 217.800003 {
				t.Fatalf("voltage metric value = %v, want 217.800003", got)
			}
		case pb.Counter != nil:
			dischargeFound = true
			if got := pb.GetCounter().GetValue(); got != 59 {
				t.Fatalf("discharge count value = %v, want 59", got)
			}
		}
	}

	if !voltageFound || !dischargeFound {
		t.Fatalf("expected both gauge and counter metrics, got voltage=%v discharge=%v", voltageFound, dischargeFound)
	}
}
