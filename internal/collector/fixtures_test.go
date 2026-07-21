package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"vertiv_exporter/internal/client"
	"vertiv_exporter/internal/config"
)

func TestACFixtureEndToEnd(t *testing.T) {
	raw := "3021,AC_1,ENP_AC_SRVII[COM]^2,Return air temperature measurement,28.600000,℃,1778314683,0,1,1,2,2;" +
		"3,Return air humidity measurement,30.400000,%,1778314683,0,1,1,2,2;" +
		"29,Power frequency,49.900002,HZ,1778314683,0,1,1,2,2;" +
		"261,Compressor output,OutputMove[1],,1778314683,0,1,1,0,5;"

	samples, err := client.ParseSamples(raw)
	if err != nil {
		t.Fatalf("ParseSamples returned error: %v", err)
	}

	col := &VertivCollector{descs: mustACDescs(t)}
	metrics := col.buildDeviceMetrics("dc-rack-01", config.Device{Name: "AC_1", Type: "ac", EquipID: 23}, samples)
	if len(metrics) != 4 {
		t.Fatalf("metric count = %d, want 4", len(metrics))
	}

	values := collectGaugeValues(t, metrics)
	if got := values["vertiv_ac_temperature_return_air_celsius"]; got != 28.6 {
		t.Fatalf("return air temp = %v, want 28.6", got)
	}
	if got := values["vertiv_ac_humidity_return_air_percent"]; got != 30.4 {
		t.Fatalf("humidity = %v, want 30.4", got)
	}
	if got := values["vertiv_ac_electrical_frequency_hz"]; got != 49.900002 {
		t.Fatalf("frequency = %v, want 49.900002", got)
	}
	if got := values["vertiv_ac_status_compressor_output"]; got != 1 {
		t.Fatalf("compressor output = %v, want 1", got)
	}
}

func TestTHDFixtureEndToEnd(t *testing.T) {
	raw := "5005,ENV_THD,THD_SENSOR^3,RACK1 Cool Aisle Top Temp,21.700001,℃,1778314638,0,1,1,2,2;" +
		"5,RACK1 Cool Aisle Hum,40.400002,%,1778314638,0,1,1,2,2;" +
		"12,RACK1 Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;" +
		"3010,RACK1 THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;" +
		"10012,High Temp Alarm Rack Count,0,,1778314638,0,1,1,0,3;"

	samples, err := client.ParseSamples(raw)
	if err != nil {
		t.Fatalf("ParseSamples returned error: %v", err)
	}

	col := &VertivCollector{thdDescs: newTHDDescs()}
	metrics := col.buildDeviceMetrics("dc-rack-01", config.Device{Name: "ENV_THD", Type: "thd", EquipID: -98}, samples)
	if len(metrics) != 5 {
		t.Fatalf("metric count = %d, want 5", len(metrics))
	}

	// Assert that THD values survive parse + routing and temperature/humidity are rounded.
	foundRoundedTemp := false
	foundRoundedHumidity := false
	for _, metric := range metrics {
		var pb dto.Metric
		if err := metric.Write(&pb); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		switch len(pb.GetLabel()) {
		case 6:
			if pb.GetGauge().GetValue() == 21.7 {
				foundRoundedTemp = true
			}
		case 5:
			if pb.GetGauge().GetValue() == 40.4 {
				foundRoundedHumidity = true
			}
		}
	}
	if !foundRoundedTemp {
		t.Fatal("did not find rounded THD temperature metric")
	}
	if !foundRoundedHumidity {
		t.Fatal("did not find rounded THD humidity metric")
	}
}

func TestUPSFixtureEndToEnd(t *testing.T) {
	raw := "491,UPS_1,ENP_UPS_ITA2[COM]^2,Phase A Input Voltage,217.800003,V,1778314661,0,1,1,2,2;" +
		"11,Output Frequency,49.970001,HZ,1778314661,0,1,1,2,2;" +
		"134,Local Phase A Output Active Power,3.110000,KW,1778314661,0,1,1,2,2;" +
		"72,Battery Current Capacity,100.000000,%,1778314661,0,1,1,2,2;" +
		"73,Battery Discharge Times,59.000000,,1778314661,0,1,1,2,2;" +
		"25,Power Supply,Utility Online[1],,1778314661,0,1,1,0,5;"

	samples, err := client.ParseSamples(raw)
	if err != nil {
		t.Fatalf("ParseSamples returned error: %v", err)
	}

	col := &VertivCollector{upsDescs: newUPSDescs()}
	metrics := col.buildDeviceMetrics("dc-rack-01", config.Device{Name: "UPS_1", Type: "ups", EquipID: 26}, samples)
	if len(metrics) != 6 {
		t.Fatalf("metric count = %d, want 6", len(metrics))
	}

	var (
		foundGauge   bool
		foundCounter bool
	)
	for _, metric := range metrics {
		var pb dto.Metric
		if err := metric.Write(&pb); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		if pb.Gauge != nil && pb.GetGauge().GetValue() == 217.800003 {
			foundGauge = true
		}
		if pb.Counter != nil && pb.GetCounter().GetValue() == 59 {
			foundCounter = true
		}
	}
	if !foundGauge {
		t.Fatal("did not find expected UPS gauge metric value")
	}
	if !foundCounter {
		t.Fatal("did not find expected UPS counter metric value")
	}
}

func mustACDescs(t *testing.T) map[int]*prometheus.Desc {
	t.Helper()
	defs, err := LoadMetricDefinitions("")
	if err != nil {
		t.Fatalf("LoadMetricDefinitions returned error: %v", err)
	}

	descs := make(map[int]*prometheus.Desc, len(defs))
	for fieldID, def := range defs {
		descs[fieldID] = prometheus.NewDesc(def.Name, def.Help, metricLabels, nil)
	}
	return descs
}

func collectGaugeValues(t *testing.T, metrics []prometheus.Metric) map[string]float64 {
	t.Helper()
	values := make(map[string]float64, len(metrics))
	for _, metric := range metrics {
		descText := metric.Desc().String()
		name := metricNameFromDesc(descText)
		var pb dto.Metric
		if err := metric.Write(&pb); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		values[name] = pb.GetGauge().GetValue()
	}
	return values
}

func metricNameFromDesc(desc string) string {
	const prefix = `Desc{fqName: "`
	start := len(prefix)
	if len(desc) <= start {
		return desc
	}
	rest := desc[start:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == '"' {
			return rest[:i]
		}
	}
	return desc
}
