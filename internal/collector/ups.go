package collector

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"vertiv_exporter/internal/client"
	"vertiv_exporter/internal/config"
)

type upsDescs struct {
	inputPhaseVoltage          *prometheus.Desc
	inputLineVoltage           *prometheus.Desc
	inputCurrent               *prometheus.Desc
	inputFrequency             *prometheus.Desc
	inputPowerFactor           *prometheus.Desc
	bypassPhaseVoltage         *prometheus.Desc
	bypassLineVoltage          *prometheus.Desc
	bypassFrequency            *prometheus.Desc
	outputPhaseVoltage         *prometheus.Desc
	outputCurrent              *prometheus.Desc
	outputFrequency            *prometheus.Desc
	outputPowerFactor          *prometheus.Desc
	outputActivePower          *prometheus.Desc
	outputApparentPower        *prometheus.Desc
	outputLoadPercent          *prometheus.Desc
	batteryVoltage             *prometheus.Desc
	batteryNegativeVoltage     *prometheus.Desc
	batteryChargeCurrent       *prometheus.Desc
	batteryDischargeCurrent    *prometheus.Desc
	batteryNegChargeCurrent    *prometheus.Desc
	batteryNegDischargeCurrent *prometheus.Desc
	batteryCapacity            *prometheus.Desc
	batteryBackupTime          *prometheus.Desc
	batteryDischargingTime     *prometheus.Desc
	batteryDischargeCount      *prometheus.Desc
	inputEnergy                *prometheus.Desc
	outputEnergy               *prometheus.Desc
	ambientTemperature         *prometheus.Desc
	runningTimeDays            *prometheus.Desc
	parallelMachineCount       *prometheus.Desc
	statusPowerSupply          *prometheus.Desc
	statusInputPower           *prometheus.Desc
	statusBattery              *prometheus.Desc
	statusBatteryNegative      *prometheus.Desc
	statusCharger              *prometheus.Desc
	statusParallelSystemPower  *prometheus.Desc
	statusInnerNetwork         *prometheus.Desc
	statusCommunication        *prometheus.Desc
	statusInputPhaseNumber     *prometheus.Desc
	statusOutputPhaseNumber    *prometheus.Desc
}

func (d upsDescs) all() []*prometheus.Desc {
	return []*prometheus.Desc{
		d.inputPhaseVoltage,
		d.inputLineVoltage,
		d.inputCurrent,
		d.inputFrequency,
		d.inputPowerFactor,
		d.bypassPhaseVoltage,
		d.bypassLineVoltage,
		d.bypassFrequency,
		d.outputPhaseVoltage,
		d.outputCurrent,
		d.outputFrequency,
		d.outputPowerFactor,
		d.outputActivePower,
		d.outputApparentPower,
		d.outputLoadPercent,
		d.batteryVoltage,
		d.batteryNegativeVoltage,
		d.batteryChargeCurrent,
		d.batteryDischargeCurrent,
		d.batteryNegChargeCurrent,
		d.batteryNegDischargeCurrent,
		d.batteryCapacity,
		d.batteryBackupTime,
		d.batteryDischargingTime,
		d.batteryDischargeCount,
		d.inputEnergy,
		d.outputEnergy,
		d.ambientTemperature,
		d.runningTimeDays,
		d.parallelMachineCount,
		d.statusPowerSupply,
		d.statusInputPower,
		d.statusBattery,
		d.statusBatteryNegative,
		d.statusCharger,
		d.statusParallelSystemPower,
		d.statusInnerNetwork,
		d.statusCommunication,
		d.statusInputPhaseNumber,
		d.statusOutputPhaseNumber,
	}
}

type upsMetricSpec struct {
	desc      func(upsDescs) *prometheus.Desc
	valueType prometheus.ValueType
	extra     []string
}

func newUPSDescs() upsDescs {
	base := []string{"instance", "device", "equip_id"}
	withPhase := []string{"instance", "device", "equip_id", "phase"}
	withLine := []string{"instance", "device", "equip_id", "line"}
	withScopePhase := []string{"instance", "device", "equip_id", "scope", "phase"}

	return upsDescs{
		inputPhaseVoltage:          prometheus.NewDesc("vertiv_ups_input_phase_voltage_volts", "UPS input phase voltage", withPhase, nil),
		inputLineVoltage:           prometheus.NewDesc("vertiv_ups_input_line_voltage_volts", "UPS input line voltage", withLine, nil),
		inputCurrent:               prometheus.NewDesc("vertiv_ups_input_current_amperes", "UPS input current", withPhase, nil),
		inputFrequency:             prometheus.NewDesc("vertiv_ups_input_frequency_hz", "UPS input frequency", base, nil),
		inputPowerFactor:           prometheus.NewDesc("vertiv_ups_input_power_factor", "UPS input power factor", withPhase, nil),
		bypassPhaseVoltage:         prometheus.NewDesc("vertiv_ups_bypass_phase_voltage_volts", "UPS bypass phase voltage", withPhase, nil),
		bypassLineVoltage:          prometheus.NewDesc("vertiv_ups_bypass_line_voltage_volts", "UPS bypass line voltage", withLine, nil),
		bypassFrequency:            prometheus.NewDesc("vertiv_ups_bypass_frequency_hz", "UPS bypass frequency", base, nil),
		outputPhaseVoltage:         prometheus.NewDesc("vertiv_ups_output_phase_voltage_volts", "UPS output phase voltage", withPhase, nil),
		outputCurrent:              prometheus.NewDesc("vertiv_ups_output_current_amperes", "UPS output current", withPhase, nil),
		outputFrequency:            prometheus.NewDesc("vertiv_ups_output_frequency_hz", "UPS output frequency", base, nil),
		outputPowerFactor:          prometheus.NewDesc("vertiv_ups_output_power_factor", "UPS output power factor", withPhase, nil),
		outputActivePower:          prometheus.NewDesc("vertiv_ups_output_active_power_kilowatts", "UPS output active power", withScopePhase, nil),
		outputApparentPower:        prometheus.NewDesc("vertiv_ups_output_apparent_power_kva", "UPS output apparent power", withScopePhase, nil),
		outputLoadPercent:          prometheus.NewDesc("vertiv_ups_output_load_percent", "UPS output load percentage per phase", withPhase, nil),
		batteryVoltage:             prometheus.NewDesc("vertiv_ups_battery_voltage_volts", "UPS battery positive group voltage", base, nil),
		batteryNegativeVoltage:     prometheus.NewDesc("vertiv_ups_battery_negative_voltage_volts", "UPS battery negative group voltage", base, nil),
		batteryChargeCurrent:       prometheus.NewDesc("vertiv_ups_battery_charge_current_amperes", "UPS battery positive group charge current", base, nil),
		batteryDischargeCurrent:    prometheus.NewDesc("vertiv_ups_battery_discharge_current_amperes", "UPS battery positive group discharge current", base, nil),
		batteryNegChargeCurrent:    prometheus.NewDesc("vertiv_ups_battery_negative_charge_current_amperes", "UPS battery negative group charge current", base, nil),
		batteryNegDischargeCurrent: prometheus.NewDesc("vertiv_ups_battery_negative_discharge_current_amperes", "UPS battery negative group discharge current", base, nil),
		batteryCapacity:            prometheus.NewDesc("vertiv_ups_battery_capacity_percent", "UPS battery state of charge", base, nil),
		batteryBackupTime:          prometheus.NewDesc("vertiv_ups_battery_backup_time_minutes", "Estimated UPS battery backup time", base, nil),
		batteryDischargingTime:     prometheus.NewDesc("vertiv_ups_battery_discharging_time_seconds", "UPS cumulative battery discharging time", base, nil),
		batteryDischargeCount:      prometheus.NewDesc("vertiv_ups_battery_discharge_count_total", "UPS cumulative battery discharge count", base, nil),
		inputEnergy:                prometheus.NewDesc("vertiv_ups_input_energy_kwh_total", "Total UPS input energy", base, nil),
		outputEnergy:               prometheus.NewDesc("vertiv_ups_output_energy_kwh_total", "Total UPS output energy", base, nil),
		ambientTemperature:         prometheus.NewDesc("vertiv_ups_ambient_temperature_celsius", "UPS ambient temperature", base, nil),
		runningTimeDays:            prometheus.NewDesc("vertiv_ups_running_time_days", "UPS running time in days", base, nil),
		parallelMachineCount:       prometheus.NewDesc("vertiv_ups_parallel_machine_count", "UPS parallel machine count", base, nil),
		statusPowerSupply:          prometheus.NewDesc("vertiv_ups_status_power_supply", "UPS power supply status (1=Utility Online)", base, nil),
		statusInputPower:           prometheus.NewDesc("vertiv_ups_status_input_power", "UPS input power status", base, nil),
		statusBattery:              prometheus.NewDesc("vertiv_ups_status_battery", "UPS battery status", base, nil),
		statusBatteryNegative:      prometheus.NewDesc("vertiv_ups_status_battery_negative_group", "UPS negative battery group status", base, nil),
		statusCharger:              prometheus.NewDesc("vertiv_ups_status_charger", "UPS charger status", base, nil),
		statusParallelSystemPower:  prometheus.NewDesc("vertiv_ups_status_parallel_system_power", "UPS parallel system power state", base, nil),
		statusInnerNetwork:         prometheus.NewDesc("vertiv_ups_status_inner_network", "UPS inner network connection status", base, nil),
		statusCommunication:        prometheus.NewDesc("vertiv_ups_status_communication", "UPS communication status", base, nil),
		statusInputPhaseNumber:     prometheus.NewDesc("vertiv_ups_status_input_phase_number", "UPS input phase number", base, nil),
		statusOutputPhaseNumber:    prometheus.NewDesc("vertiv_ups_status_output_phase_number", "UPS output phase number", base, nil),
	}
}

var upsSpecs = map[int]upsMetricSpec{
	2:   {desc: func(d upsDescs) *prometheus.Desc { return d.inputPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	3:   {desc: func(d upsDescs) *prometheus.Desc { return d.inputPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	4:   {desc: func(d upsDescs) *prometheus.Desc { return d.inputPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	16:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputLineVoltage }, valueType: prometheus.GaugeValue, extra: []string{"AB"}},
	69:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputLineVoltage }, valueType: prometheus.GaugeValue, extra: []string{"BC"}},
	63:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputLineVoltage }, valueType: prometheus.GaugeValue, extra: []string{"CA"}},
	124: {desc: func(d upsDescs) *prometheus.Desc { return d.inputCurrent }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	125: {desc: func(d upsDescs) *prometheus.Desc { return d.inputCurrent }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	126: {desc: func(d upsDescs) *prometheus.Desc { return d.inputCurrent }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	12:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputFrequency }, valueType: prometheus.GaugeValue},
	23:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputPowerFactor }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	71:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputPowerFactor }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	77:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputPowerFactor }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	143: {desc: func(d upsDescs) *prometheus.Desc { return d.bypassPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	144: {desc: func(d upsDescs) *prometheus.Desc { return d.bypassPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	145: {desc: func(d upsDescs) *prometheus.Desc { return d.bypassPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	29:  {desc: func(d upsDescs) *prometheus.Desc { return d.bypassLineVoltage }, valueType: prometheus.GaugeValue, extra: []string{"AB"}},
	30:  {desc: func(d upsDescs) *prometheus.Desc { return d.bypassLineVoltage }, valueType: prometheus.GaugeValue, extra: []string{"BC"}},
	31:  {desc: func(d upsDescs) *prometheus.Desc { return d.bypassLineVoltage }, valueType: prometheus.GaugeValue, extra: []string{"CA"}},
	146: {desc: func(d upsDescs) *prometheus.Desc { return d.bypassFrequency }, valueType: prometheus.GaugeValue},
	5:   {desc: func(d upsDescs) *prometheus.Desc { return d.outputPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	6:   {desc: func(d upsDescs) *prometheus.Desc { return d.outputPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	7:   {desc: func(d upsDescs) *prometheus.Desc { return d.outputPhaseVoltage }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	8:   {desc: func(d upsDescs) *prometheus.Desc { return d.outputCurrent }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	9:   {desc: func(d upsDescs) *prometheus.Desc { return d.outputCurrent }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	10:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputCurrent }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	11:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputFrequency }, valueType: prometheus.GaugeValue},
	35:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputPowerFactor }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	36:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputPowerFactor }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	37:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputPowerFactor }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	134: {desc: func(d upsDescs) *prometheus.Desc { return d.outputActivePower }, valueType: prometheus.GaugeValue, extra: []string{"local", "A"}},
	135: {desc: func(d upsDescs) *prometheus.Desc { return d.outputActivePower }, valueType: prometheus.GaugeValue, extra: []string{"local", "B"}},
	136: {desc: func(d upsDescs) *prometheus.Desc { return d.outputActivePower }, valueType: prometheus.GaugeValue, extra: []string{"local", "C"}},
	54:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputActivePower }, valueType: prometheus.GaugeValue, extra: []string{"system", "A"}},
	55:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputActivePower }, valueType: prometheus.GaugeValue, extra: []string{"system", "B"}},
	56:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputActivePower }, valueType: prometheus.GaugeValue, extra: []string{"system", "C"}},
	44:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputApparentPower }, valueType: prometheus.GaugeValue, extra: []string{"local", "A"}},
	45:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputApparentPower }, valueType: prometheus.GaugeValue, extra: []string{"local", "B"}},
	46:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputApparentPower }, valueType: prometheus.GaugeValue, extra: []string{"local", "C"}},
	57:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputApparentPower }, valueType: prometheus.GaugeValue, extra: []string{"system", "A"}},
	58:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputApparentPower }, valueType: prometheus.GaugeValue, extra: []string{"system", "B"}},
	59:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputApparentPower }, valueType: prometheus.GaugeValue, extra: []string{"system", "C"}},
	13:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputLoadPercent }, valueType: prometheus.GaugeValue, extra: []string{"A"}},
	14:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputLoadPercent }, valueType: prometheus.GaugeValue, extra: []string{"B"}},
	15:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputLoadPercent }, valueType: prometheus.GaugeValue, extra: []string{"C"}},
	18:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryVoltage }, valueType: prometheus.GaugeValue},
	66:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryNegativeVoltage }, valueType: prometheus.GaugeValue},
	64:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryChargeCurrent }, valueType: prometheus.GaugeValue},
	65:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryDischargeCurrent }, valueType: prometheus.GaugeValue},
	67:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryNegChargeCurrent }, valueType: prometheus.GaugeValue},
	68:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryNegDischargeCurrent }, valueType: prometheus.GaugeValue},
	72:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryCapacity }, valueType: prometheus.GaugeValue},
	17:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryBackupTime }, valueType: prometheus.GaugeValue},
	167: {desc: func(d upsDescs) *prometheus.Desc { return d.batteryDischargingTime }, valueType: prometheus.GaugeValue},
	73:  {desc: func(d upsDescs) *prometheus.Desc { return d.batteryDischargeCount }, valueType: prometheus.CounterValue},
	75:  {desc: func(d upsDescs) *prometheus.Desc { return d.inputEnergy }, valueType: prometheus.CounterValue},
	76:  {desc: func(d upsDescs) *prometheus.Desc { return d.outputEnergy }, valueType: prometheus.CounterValue},
	24:  {desc: func(d upsDescs) *prometheus.Desc { return d.ambientTemperature }, valueType: prometheus.GaugeValue},
	62:  {desc: func(d upsDescs) *prometheus.Desc { return d.runningTimeDays }, valueType: prometheus.GaugeValue},
	60:  {desc: func(d upsDescs) *prometheus.Desc { return d.parallelMachineCount }, valueType: prometheus.GaugeValue},
	25:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusPowerSupply }, valueType: prometheus.GaugeValue},
	79:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusInputPower }, valueType: prometheus.GaugeValue},
	27:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusBattery }, valueType: prometheus.GaugeValue},
	81:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusBatteryNegative }, valueType: prometheus.GaugeValue},
	82:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusCharger }, valueType: prometheus.GaugeValue},
	83:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusParallelSystemPower }, valueType: prometheus.GaugeValue},
	84:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusInnerNetwork }, valueType: prometheus.GaugeValue},
	456: {desc: func(d upsDescs) *prometheus.Desc { return d.statusCommunication }, valueType: prometheus.GaugeValue},
	49:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusInputPhaseNumber }, valueType: prometheus.GaugeValue},
	34:  {desc: func(d upsDescs) *prometheus.Desc { return d.statusOutputPhaseNumber }, valueType: prometheus.GaugeValue},
}

func isUPSDevice(device config.Device) bool {
	if strings.EqualFold(device.Type, "ups") {
		return true
	}
	return strings.Contains(strings.ToUpper(device.Name), "UPS")
}

func buildUPSMetrics(descs upsDescs, instance string, device config.Device, samples map[int]client.Sample) []prometheus.Metric {
	base := []string{instance, device.Name, strconv.Itoa(device.EquipID)}
	metrics := make([]prometheus.Metric, 0, len(samples))

	for fieldID, sample := range samples {
		spec, ok := upsSpecs[fieldID]
		if !ok {
			continue
		}

		labels := append(append([]string{}, base...), spec.extra...)
		metrics = append(metrics, prometheus.MustNewConstMetric(
			spec.desc(descs),
			spec.valueType,
			sample.Value,
			labels...,
		))
	}

	return metrics
}
