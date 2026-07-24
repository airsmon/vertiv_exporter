# Vertiv Exporter Grafana Dashboard

`vertiv_grafana_dashboard.json` 是可直接导入 Grafana 的标准 dashboard JSON，默认兼容 Grafana 10 及以上版本。

## 导入

1. 在 Grafana 中打开 **Dashboards → New → Import**。
2. 上传 [`vertiv_grafana_dashboard.json`](vertiv_grafana_dashboard.json)。
3. 在导入页为 `DS_PROMETHEUS` 选择实际的 Prometheus 数据源。
4. 导入后通过顶部的 `job`、`instance` 和 `device` 变量切换采集任务、Exporter 实例和 AC 设备。

Dashboard 不包含硬编码的数据源 UID。所有面板均使用 `${DS_PROMETHEUS}`，PromQL 同时带有 `job` / `instance` 过滤条件。

> `instance` 是 Prometheus 的抓取目标标签。Exporter 自身也导出了同名标签；在 Prometheus 默认 `honor_labels: false` 配置下，Exporter 原始标签会被重命名为 `exported_instance`。需要按 Vertiv 目标名称筛选时，请参考项目根目录 `REVIEW.md` 中的标签迁移建议。

## 面板与指标

### Exporter 健康

| 面板 | 指标 |
| --- | --- |
| Exporter / Target Health | `vertiv_exporter_up` |
| Scrape Duration p95 | `vertiv_exporter_scrape_duration_seconds_bucket` |
| Scrape Failures in Range | `vertiv_exporter_scrape_failures_total` |

### ENV_THD

| 面板 | 指标 |
| --- | --- |
| THD Temperature | `vertiv_thd_temperature_celsius` |
| THD Humidity | `vertiv_thd_humidity_percent` |
| THD Door Status | `vertiv_thd_door_status` |
| THD Communication | `vertiv_thd_comm_status` |
| High-temp Racks | `vertiv_thd_high_temp_alarm_rack_count` |
| THD Rack IDs | `vertiv_thd_rack_id` |

### UPS

| 面板 | 指标 |
| --- | --- |
| UPS Input & Bypass Voltage | `vertiv_ups_input_phase_voltage_volts`, `vertiv_ups_input_line_voltage_volts`, `vertiv_ups_bypass_phase_voltage_volts`, `vertiv_ups_bypass_line_voltage_volts` |
| UPS Output Voltage | `vertiv_ups_output_phase_voltage_volts` |
| UPS Phase Current | `vertiv_ups_input_current_amperes`, `vertiv_ups_output_current_amperes` |
| UPS Frequency | `vertiv_ups_input_frequency_hz`, `vertiv_ups_output_frequency_hz`, `vertiv_ups_bypass_frequency_hz` |
| UPS Output Power | `vertiv_ups_output_active_power_kilowatts`, `vertiv_ups_output_apparent_power_kva` |
| UPS Load & Power Factor | `vertiv_ups_output_load_percent`, `vertiv_ups_input_power_factor`, `vertiv_ups_output_power_factor` |
| UPS Battery Electrical Values | `vertiv_ups_battery_voltage_volts`, `vertiv_ups_battery_negative_voltage_volts`, `vertiv_ups_battery_charge_current_amperes`, `vertiv_ups_battery_discharge_current_amperes`, `vertiv_ups_battery_negative_charge_current_amperes`, `vertiv_ups_battery_negative_discharge_current_amperes` |
| UPS Battery, Energy & Runtime | `vertiv_ups_battery_capacity_percent`, `vertiv_ups_battery_backup_time_minutes`, `vertiv_ups_battery_discharging_time_seconds`, `vertiv_ups_battery_discharge_count_total`, `vertiv_ups_input_energy_kwh_total`, `vertiv_ups_output_energy_kwh_total`, `vertiv_ups_ambient_temperature_celsius`, `vertiv_ups_running_time_days`, `vertiv_ups_parallel_machine_count` |
| UPS Status Values | 所有 `vertiv_ups_status_*` 指标 |

### AC

| 面板 | 指标 |
| --- | --- |
| AC Temperature Metrics | 所有 `vertiv_ac_temperature_*_celsius` 指标 |
| AC Humidity Metrics | 所有 `vertiv_ac_humidity_*_percent` 指标 |
| AC Pressure Metrics | 所有 `vertiv_ac_pressure_*_bar` 指标 |
| AC Electrical Metrics | 所有 `vertiv_ac_electrical_*` 指标 |
| AC Compressor, Fan & EEV | `vertiv_ac_compressor_*`、`vertiv_ac_fan_*`、`vertiv_ac_eev_*`（累计值除外） |
| AC Runtime & Maintenance | AC 压缩机、风机、泵、加湿器、电加热累计值，以及除湿和过滤器维护指标 |
| AC Status Values | 所有 `vertiv_ac_status_*` 指标 |
| AC Alarm Attributes | 所有 `vertiv_ac_alarm_attr_*` 指标 |
| AC System Values | 所有 `vertiv_ac_system_*` 指标 |
| AC Raw Signals | `vertiv_ac_signal_value` |

完整指标类型、单位和 Label 以项目根目录 `README.md` 的「指标说明」为准。
