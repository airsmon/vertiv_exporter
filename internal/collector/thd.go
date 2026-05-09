package collector

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"vertiv_exporter/internal/client"
	"vertiv_exporter/internal/config"
)

var (
	thdTempPattern = regexp.MustCompile(`^(RACK(?:[A-Z0-9]+)?(?:\s+[A-Z0-9]+)*)\s+(Cool|Hot)\s+Aisle\s+(Top|Middle|Bottom)\s+Temp$`)
	thdHumPattern  = regexp.MustCompile(`^(RACK(?:[A-Z0-9]+)?(?:\s+[A-Z0-9]+)*)\s+(Cool|Hot)\s+Aisle\s+Hum$`)
	thdDoorPattern = regexp.MustCompile(`^(RACK(?:[A-Z0-9]+)?(?:\s+[A-Z0-9]+)*)\s+(Cool|Hot)\s+Aisle\s+Door Status$`)
	thdCommPattern = regexp.MustCompile(`^(RACK(?:[A-Z0-9]+)?(?:\s+[A-Z0-9]+)*)\s+THD Comm Status$`)
)

type thdDescs struct {
	temperature        *prometheus.Desc
	humidity           *prometheus.Desc
	doorStatus         *prometheus.Desc
	commStatus         *prometheus.Desc
	rackID             *prometheus.Desc
	highTempAlarmCount *prometheus.Desc
}

func (d thdDescs) all() []*prometheus.Desc {
	return []*prometheus.Desc{
		d.temperature,
		d.humidity,
		d.doorStatus,
		d.commStatus,
		d.rackID,
		d.highTempAlarmCount,
	}
}

func newTHDDescs() thdDescs {
	return thdDescs{
		temperature: prometheus.NewDesc(
			"vertiv_thd_temperature_celsius",
			"Aisle temperature measurement",
			[]string{"instance", "device", "equip_id", "rack", "aisle", "position"},
			nil,
		),
		humidity: prometheus.NewDesc(
			"vertiv_thd_humidity_percent",
			"Aisle humidity measurement",
			[]string{"instance", "device", "equip_id", "rack", "aisle"},
			nil,
		),
		doorStatus: prometheus.NewDesc(
			"vertiv_thd_door_status",
			"Aisle door status (0=Normal, 1=Open/Abnormal)",
			[]string{"instance", "device", "equip_id", "rack", "aisle"},
			nil,
		),
		commStatus: prometheus.NewDesc(
			"vertiv_thd_comm_status",
			"THD sensor communication status (0=Normal, 1=Fault)",
			[]string{"instance", "device", "equip_id", "rack"},
			nil,
		),
		rackID: prometheus.NewDesc(
			"vertiv_thd_rack_id",
			"Rack identifier mapping from the Vertiv THD subsystem",
			[]string{"instance", "device", "equip_id", "rack"},
			nil,
		),
		highTempAlarmCount: prometheus.NewDesc(
			"vertiv_thd_high_temp_alarm_rack_count",
			"Number of racks with active high temperature alarm",
			[]string{"instance", "device", "equip_id"},
			nil,
		),
	}
}

func isTHDDevice(device config.Device) bool {
	if strings.EqualFold(device.Type, "thd") {
		return true
	}
	name := strings.ToUpper(device.Name)
	return device.EquipID == 5005 || strings.Contains(name, "THD")
}

func buildTHDMetrics(descs thdDescs, instance string, device config.Device, samples map[int]client.Sample) []prometheus.Metric {
	base := []string{instance, device.Name, strconv.Itoa(device.EquipID)}
	metrics := make([]prometheus.Metric, 0, len(samples))

	for fieldID, sample := range samples {
		name := strings.TrimSpace(sample.Name)

		switch {
		case name == "High Temp Alarm Rack Count":
			metrics = append(metrics, prometheus.MustNewConstMetric(
				descs.highTempAlarmCount,
				prometheus.GaugeValue,
				sample.Value,
				base...,
			))
		case fieldID >= 10000 && strings.HasPrefix(name, "RACK"):
			metrics = append(metrics, prometheus.MustNewConstMetric(
				descs.rackID,
				prometheus.GaugeValue,
				sample.Value,
				append(base, normalizeRack(name))...,
			))
		case thdTempPattern.MatchString(name):
			m := thdTempPattern.FindStringSubmatch(name)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				descs.temperature,
				prometheus.GaugeValue,
				roundTo2(sample.Value),
				append(base, normalizeRack(m[1]), normalizeAisle(m[2]), normalizePosition(m[3]))...,
			))
		case thdHumPattern.MatchString(name):
			m := thdHumPattern.FindStringSubmatch(name)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				descs.humidity,
				prometheus.GaugeValue,
				roundTo2(sample.Value),
				append(base, normalizeRack(m[1]), normalizeAisle(m[2]))...,
			))
		case thdDoorPattern.MatchString(name):
			m := thdDoorPattern.FindStringSubmatch(name)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				descs.doorStatus,
				prometheus.GaugeValue,
				sample.Value,
				append(base, normalizeRack(m[1]), normalizeAisle(m[2]))...,
			))
		case thdCommPattern.MatchString(name):
			m := thdCommPattern.FindStringSubmatch(name)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				descs.commStatus,
				prometheus.GaugeValue,
				sample.Value,
				append(base, normalizeRack(m[1]))...,
			))
		}
	}

	return metrics
}

func normalizeRack(raw string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(raw)), " ", "_")
}

func normalizeAisle(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizePosition(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func roundTo2(v float64) float64 {
	return math.Round(v*100) / 100
}
