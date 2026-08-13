
# Vertiv RDU-A G2智能监控单元

<img alt="FireShot Capture 061 - 上海道客网络科技有限公司【尚浦中心】 - RDU-A G2智能监控单元 - Dashboards - Grafana_ -  grafana infra daocloud io" src="https://github.com/user-attachments/assets/0b399aa5-1e7c-450d-b201-b91e741485b6" width="100%" />

Prometheus exporter for Vertiv devices through the Vertiv CGI interface.

Current supported device families:

- `ac`: Vertiv AC / precision cooling devices
- `thd`: `ENV_THD` rack aisle temperature and humidity sensors
- `ups`: Vertiv UPS devices

## Features

- Scrapes Vertiv CGI endpoint `p05_equip_sample.cgi`
- Maintains login session and keepalive automatically
- Exposes Prometheus metrics on `/metrics`
- Supports multiple targets and multiple devices per target
- Supports device-specific metric shaping:
  - AC: field-id to metric mapping
  - THD: label-based metrics with `rack`, `aisle`, `position`
  - UPS: label-based metrics with `phase`, `line`, `scope`

## Requirements

- Go `1.21+`
- Network access from the exporter host to the Vertiv web interface

## Project Layout

```text
Dockerfile                  Multi-stage, non-root container image
.dockerignore               Minimal production build context allowlist
.github/                    CI, GHCR publishing, dependency updates, and templates
charts/                     Helm chart with optional Prometheus Operator resources
.config.yaml                Example exporter configuration
grafana/                    Standard Grafana dashboard and import guide
packaging/systemd/          Hardened systemd service unit
cmd/vertiv_exporter/        CLI entrypoint and HTTP server
internal/client/            Login, keepalive, CGI fetching, response parsing
internal/collector/         Collector, built-in AC metadata, and THD/UPS mappings
internal/config/            YAML config loader
```

## Configuration

Example config:

```yaml
exporter:
  listen_address: ":9101"
  metrics_path: "/metrics"
  scrape_timeout: 10s
  metrics_file: ""
  debug_response: false

targets:
  - name: "dc-rack-01"
    host: "https://vertiv.example.invalid"
    username: "CHANGE_ME_USERNAME"
    password: "CHANGE_ME_PASSWORD"
    tls_skip_verify: true
    devices:
      - name: "AC_1"
        type: "ac"
        equip_id: 23
      - name: "ENV_THD"
        type: "thd"
        equip_id: -98
      - name: "UPS_1"
        type: "ups"
        equip_id: 26
```

### Config Fields

- `exporter.listen_address`: HTTP listen address; defaults to `:9101`
- `exporter.metrics_path`: Prometheus endpoint; defaults to `/metrics`
- `exporter.scrape_timeout`: timeout for one collector run across all targets; defaults to `10s`
- `exporter.metrics_file`: optional Markdown file that overrides the built-in AC field mapping; an empty or unreadable path falls back to the built-in mapping
- `exporter.debug_response`: when `true`, parse failures include the full CGI response body in logs
- `target.name`: value used as the Prometheus `target` label
- `host`: Vertiv web base URL
- `username` / `password`: plain-text login values; the exporter encodes them automatically before calling `login.cgi`
- `VERTIV_USERNAME` / `VERTIV_PASSWORD`: optional environment variables that override the YAML credentials for every target; they must be set together to non-empty values and are intended for Docker or Kubernetes Secret injection
- `tls_skip_verify`: useful for self-signed Vertiv HTTPS endpoints
- `device.type`: device-family hint; use `ac`, `thd`, or `ups`
- `device.equip_id`: CGI request `_equipId` value used for that device

`metrics_path` must be an absolute non-root path without URL query/fragment, whitespace, repeated slashes, or `.`/`..` segments; one trailing slash is allowed. `scrape_timeout` must be positive, and target names must be unique.

### Device Type Notes

- `ac` devices usually use positive `equip_id` values such as `23`, `24`
- `thd` devices may use request IDs like `-98`
- `ups` devices use their own request `equip_id`, for example `26`; this can differ from an internal device code such as `491` in the response body

Mapping selection follows the implementation's precedence: `type: thd`, `equip_id: 5005`, or a name containing `THD` selects THD first; otherwise `type: ups` or a name containing `UPS` selects UPS; all remaining devices use AC. Use conventional names and set `type` explicitly to make the intended family clear.

The IDs above are examples from one tested installation. Confirm the CGI request `_equipId` values for each target before deployment.

### Credential Safety

- The YAML file may contain plain-text credentials. Keep `config.yaml` out of version control. For direct execution, restrict it with `chmod 600 config.yaml`; for systemd, use the `root:vertiv_exporter` ownership and `0640` mode documented below.
- For Kubernetes, omit `username` and `password` from the ConfigMap and inject non-empty `VERTIV_USERNAME` and `VERTIV_PASSWORD` values from a Secret.
- The environment-variable override applies one credential pair to every target. When targets require different credentials, mount the complete `config.yaml` from a Secret instead.
- Mount credentials through a Docker/Kubernetes secret rather than baking them into an image.
- Enable `tls_skip_verify` only for trusted internal endpoints whose certificate cannot be validated normally.
- Keep `debug_response` disabled during normal operation because parse errors may include complete device responses.

## Login Behavior

The exporter logs in with the same payload shape captured from the browser:

- `user_name`
- `user_password`
- `lan=en`
- `op_Type=1`
- `rand_code=0`
- `tokenID=$[$ID_TOKEN_ID]`
- `validateValue=0`

Before sending the request, the exporter automatically encodes `username` and `password` using the Vertiv-compatible scheme observed in browser traffic:

- if the value is shorter than 9 bytes, it is NUL-padded to 9 bytes
- then it is encoded with base64 without `=` padding

Session keepalive uses `main_page_polling.cgi`.

## Metric Sources

AC metric metadata is defined directly in [default_metrics.go](internal/collector/default_metrics.go). THD and UPS mappings are implemented in [thd.go](internal/collector/thd.go) and [ups.go](internal/collector/ups.go).

You can point `exporter.metrics_file` at a custom Markdown file when different AC field mappings are required. Without an override, the binary uses its built-in Go definitions and has no runtime documentation-file dependency.

Each custom mapping row must contain the Prometheus metric name, numeric field ID, and help text:

```markdown
| `vertiv_ac_temperature_return_air_celsius` | 2 | Return air temperature measurement |
```

## 指标说明

以下清单以当前代码的默认内置定义为准。所有设备指标都包含 `target`、`device`、`equip_id`；其中 `target` 是配置中的 Vertiv target 名称。若配置了 `exporter.metrics_file`，其中的 AC 字段映射会替换下列 181 个内置语义 AC 指标，`vertiv_ac_signal_value` 不受影响。

### ENV_THD（环境阈值）

共 6 个指标。温度和湿度在导出前保留两位小数；机柜、通道和位置 Label 会随设备返回的传感器拓扑动态变化。

| 指标名 | 类型（Gauge/Counter） | 单位 | Label | 说明 |
| --- | --- | --- | --- | --- |
| `vertiv_thd_temperature_celsius` | Gauge | °C | `target`, `device`, `equip_id`, `rack`, `aisle`, `position` | 通道温度；`rack`、`aisle`、`position` 从设备返回的传感器名称动态解析。 |
| `vertiv_thd_humidity_percent` | Gauge | % | `target`, `device`, `equip_id`, `rack`, `aisle` | 通道湿度；`rack`、`aisle` 从设备返回的传感器名称动态解析。 |
| `vertiv_thd_door_status` | Gauge | — | `target`, `device`, `equip_id`, `rack`, `aisle` | 通道门状态（0=正常，1=打开/异常）；`rack`、`aisle` 为动态 Label。 |
| `vertiv_thd_comm_status` | Gauge | — | `target`, `device`, `equip_id`, `rack` | THD 传感器通信状态（0=正常，1=故障）；`rack` 为动态 Label。 |
| `vertiv_thd_rack_id` | Gauge | — | `target`, `device`, `equip_id`, `rack` | Vertiv THD 子系统的机柜标识映射；`rack` 为动态 Label。 |
| `vertiv_thd_high_temp_alarm_rack_count` | Gauge | 个 | `target`, `device`, `equip_id` | 当前存在高温告警的机柜数量。 |

### UPS

共 40 个指标。`phase` 的值为 `A/B/C`，`line` 的值为 `AB/BC/CA`，`scope` 的值为 `local/system`；这些 Label 由字段 ID 映射生成。

| 指标名 | 类型（Gauge/Counter） | 单位 | Label | 说明 |
| --- | --- | --- | --- | --- |
| `vertiv_ups_input_phase_voltage_volts` | Gauge | V | `target`, `device`, `equip_id`, `phase` | UPS input phase voltage |
| `vertiv_ups_input_line_voltage_volts` | Gauge | V | `target`, `device`, `equip_id`, `line` | UPS input line voltage |
| `vertiv_ups_input_current_amperes` | Gauge | A | `target`, `device`, `equip_id`, `phase` | UPS input current |
| `vertiv_ups_input_frequency_hz` | Gauge | Hz | `target`, `device`, `equip_id` | UPS input frequency |
| `vertiv_ups_input_power_factor` | Gauge | — | `target`, `device`, `equip_id`, `phase` | UPS input power factor |
| `vertiv_ups_bypass_phase_voltage_volts` | Gauge | V | `target`, `device`, `equip_id`, `phase` | UPS bypass phase voltage |
| `vertiv_ups_bypass_line_voltage_volts` | Gauge | V | `target`, `device`, `equip_id`, `line` | UPS bypass line voltage |
| `vertiv_ups_bypass_frequency_hz` | Gauge | Hz | `target`, `device`, `equip_id` | UPS bypass frequency |
| `vertiv_ups_output_phase_voltage_volts` | Gauge | V | `target`, `device`, `equip_id`, `phase` | UPS output phase voltage |
| `vertiv_ups_output_current_amperes` | Gauge | A | `target`, `device`, `equip_id`, `phase` | UPS output current |
| `vertiv_ups_output_frequency_hz` | Gauge | Hz | `target`, `device`, `equip_id` | UPS output frequency |
| `vertiv_ups_output_power_factor` | Gauge | — | `target`, `device`, `equip_id`, `phase` | UPS output power factor |
| `vertiv_ups_output_active_power_kilowatts` | Gauge | kW | `target`, `device`, `equip_id`, `scope`, `phase` | UPS output active power |
| `vertiv_ups_output_apparent_power_kva` | Gauge | kVA | `target`, `device`, `equip_id`, `scope`, `phase` | UPS output apparent power |
| `vertiv_ups_output_load_percent` | Gauge | % | `target`, `device`, `equip_id`, `phase` | UPS output load percentage per phase |
| `vertiv_ups_battery_voltage_volts` | Gauge | V | `target`, `device`, `equip_id` | UPS battery positive group voltage |
| `vertiv_ups_battery_negative_voltage_volts` | Gauge | V | `target`, `device`, `equip_id` | UPS battery negative group voltage |
| `vertiv_ups_battery_charge_current_amperes` | Gauge | A | `target`, `device`, `equip_id` | UPS battery positive group charge current |
| `vertiv_ups_battery_discharge_current_amperes` | Gauge | A | `target`, `device`, `equip_id` | UPS battery positive group discharge current |
| `vertiv_ups_battery_negative_charge_current_amperes` | Gauge | A | `target`, `device`, `equip_id` | UPS battery negative group charge current |
| `vertiv_ups_battery_negative_discharge_current_amperes` | Gauge | A | `target`, `device`, `equip_id` | UPS battery negative group discharge current |
| `vertiv_ups_battery_capacity_percent` | Gauge | % | `target`, `device`, `equip_id` | UPS battery state of charge |
| `vertiv_ups_battery_backup_time_minutes` | Gauge | min | `target`, `device`, `equip_id` | Estimated UPS battery backup time |
| `vertiv_ups_battery_discharging_time_seconds` | Gauge | s | `target`, `device`, `equip_id` | UPS cumulative battery discharging time |
| `vertiv_ups_battery_discharge_count_total` | Counter | 次 | `target`, `device`, `equip_id` | UPS cumulative battery discharge count |
| `vertiv_ups_input_energy_kwh_total` | Counter | kWh | `target`, `device`, `equip_id` | Total UPS input energy |
| `vertiv_ups_output_energy_kwh_total` | Counter | kWh | `target`, `device`, `equip_id` | Total UPS output energy |
| `vertiv_ups_ambient_temperature_celsius` | Gauge | °C | `target`, `device`, `equip_id` | UPS ambient temperature |
| `vertiv_ups_running_time_days` | Gauge | d | `target`, `device`, `equip_id` | UPS running time in days |
| `vertiv_ups_parallel_machine_count` | Gauge | 个 | `target`, `device`, `equip_id` | UPS parallel machine count |
| `vertiv_ups_status_power_supply` | Gauge | — | `target`, `device`, `equip_id` | UPS power supply status (1=Utility Online) |
| `vertiv_ups_status_input_power` | Gauge | — | `target`, `device`, `equip_id` | UPS input power status |
| `vertiv_ups_status_battery` | Gauge | — | `target`, `device`, `equip_id` | UPS battery status |
| `vertiv_ups_status_battery_negative_group` | Gauge | — | `target`, `device`, `equip_id` | UPS negative battery group status |
| `vertiv_ups_status_charger` | Gauge | — | `target`, `device`, `equip_id` | UPS charger status |
| `vertiv_ups_status_parallel_system_power` | Gauge | — | `target`, `device`, `equip_id` | UPS parallel system power state |
| `vertiv_ups_status_inner_network` | Gauge | — | `target`, `device`, `equip_id` | UPS inner network connection status |
| `vertiv_ups_status_communication` | Gauge | — | `target`, `device`, `equip_id` | UPS communication status |
| `vertiv_ups_status_input_phase_number` | Gauge | — | `target`, `device`, `equip_id` | UPS input phase number |
| `vertiv_ups_status_output_phase_number` | Gauge | — | `target`, `device`, `equip_id` | UPS output phase number |

### AC

共 182 个指标（181 个内置语义指标 + 1 个原始信号指标）。内置语义指标统一使用 `target`、`device`、`equip_id`；原始信号额外使用动态 `signal_name` 和 `occurrence`。当前实现将所有 AC 指标按 Gauge 导出，包括名称以 `_total` 结尾的累计值。

| 指标名 | 类型（Gauge/Counter） | 单位 | Label | 说明 |
| --- | --- | --- | --- | --- |
| `vertiv_ac_temperature_return_air_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Return air temperature measurement |
| `vertiv_ac_temperature_supply_air_1_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Air supply temperature 1 measured value |
| `vertiv_ac_temperature_supply_air_2_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Air supply temperature 2 measured value |
| `vertiv_ac_temperature_supply_air_mean_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Mean temperature measurement of Supply Air |
| `vertiv_ac_temperature_supply_air_setpoint_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Supply Air temperature setting |
| `vertiv_ac_temperature_airflow_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Air temperature measurement |
| `vertiv_ac_temperature_exhaust_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Exhaust temperature measurement |
| `vertiv_ac_temperature_inspiratory_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Inspiratory temperature measurement |
| `vertiv_ac_temperature_inspiratory_evaporation_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Inspiratory evaporation temperature |
| `vertiv_ac_temperature_exhaust_condensing_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Exhaust condensing temperature |
| `vertiv_ac_temperature_inspiratory_superheat_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Inspiratory superheat |
| `vertiv_ac_temperature_exhaust_superheat_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Exhaust superheat |
| `vertiv_ac_temperature_return_air_setpoint_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Return air temperature setting |
| `vertiv_ac_temperature_remote_setpoint_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Remote temperature setting |
| `vertiv_ac_temperature_airflow_alarm_value_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Air temperature alarm value |
| `vertiv_ac_temperature_low_alarm_value_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Low temperature alarm value |
| `vertiv_ac_temperature_return_air_alarm_value_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Return air temperature alarm value |
| `vertiv_ac_temperature_airflow_loss_alarm_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Airflow loss temperature alarm value |
| `vertiv_ac_temperature_dead_zone_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Temperature dead zone |
| `vertiv_ac_temperature_dehumid_stop_diff_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Dehumidification stop temperature difference |
| `vertiv_ac_temperature_eev_superheat_setpoint_celsius` | Gauge | °C | `target`, `device`, `equip_id` | EEV superheat setting |
| `vertiv_ac_temperature_eev_close_superheat_celsius` | Gauge | °C | `target`, `device`, `equip_id` | The EEV valve closes the superheat |
| `vertiv_ac_temperature_remote_1_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Remote temperature 1 measurements |
| `vertiv_ac_temperature_remote_2_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Remote temperature 2 measurements |
| `vertiv_ac_temperature_remote_3_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Remote temperature 3 measurements |
| `vertiv_ac_temperature_remote_avg_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Remote average temperature |
| `vertiv_ac_temperature_supply_air_1_correction_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Air supply temperature 1 correction value |
| `vertiv_ac_temperature_supply_air_2_correction_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Air supply temperature 2 correction value |
| `vertiv_ac_temperature_return_air_correction_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Return air temperature correction value |
| `vertiv_ac_temperature_airflow_correction_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Airflow temperature correction value |
| `vertiv_ac_temperature_exhaust_correction_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Exhaust temperature correction value |
| `vertiv_ac_temperature_inspiratory_correction_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Inspiratory temperature correction value |
| `vertiv_ac_humidity_return_air_percent` | Gauge | % | `target`, `device`, `equip_id` | Return air humidity measurement |
| `vertiv_ac_humidity_return_air_setpoint_percent` | Gauge | % | `target`, `device`, `equip_id` | Humidity setting |
| `vertiv_ac_humidity_ratio_percent` | Gauge | % | `target`, `device`, `equip_id` | Humidity ratio |
| `vertiv_ac_humidity_supply_air_theoretical_percent` | Gauge | % | `target`, `device`, `equip_id` | Theoretical air supply humidity |
| `vertiv_ac_humidity_supply_air_current_percent` | Gauge | % | `target`, `device`, `equip_id` | Current air supply humidity |
| `vertiv_ac_humidity_dead_zone_percent` | Gauge | % | `target`, `device`, `equip_id` | Humidity dead zone |
| `vertiv_ac_humidity_return_air_high_alarm_percent` | Gauge | % | `target`, `device`, `equip_id` | Return wind high humidity alarm value |
| `vertiv_ac_humidity_return_air_low_alarm_percent` | Gauge | % | `target`, `device`, `equip_id` | Return air low warning value |
| `vertiv_ac_humidity_return_air_correction_percent` | Gauge | % | `target`, `device`, `equip_id` | Return air humidity correction value |
| `vertiv_ac_pressure_exhaust_bar` | Gauge | bar | `target`, `device`, `equip_id` | Exhaust pressure measurement |
| `vertiv_ac_pressure_inspiratory_bar` | Gauge | bar | `target`, `device`, `equip_id` | Inspiratory pressure measurement |
| `vertiv_ac_pressure_exhaust_correction_bar` | Gauge | bar | `target`, `device`, `equip_id` | Exhaust pressure correction value |
| `vertiv_ac_pressure_inspiratory_correction_bar` | Gauge | bar | `target`, `device`, `equip_id` | Inspiratory pressure correction value |
| `vertiv_ac_pressure_eev_mop_limit_bar` | Gauge | bar | `target`, `device`, `equip_id` | EEV MOP pressure limit |
| `vertiv_ac_pressure_fan_max_speed_low_point_bar` | Gauge | bar | `target`, `device`, `equip_id` | Fan maximum speed low pressure point |
| `vertiv_ac_pressure_fan_speed_low_point_bar` | Gauge | bar | `target`, `device`, `equip_id` | Fan speed low point |
| `vertiv_ac_pressure_fan_down_low_point_bar` | Gauge | bar | `target`, `device`, `equip_id` | Fan down the low pressure point |
| `vertiv_ac_pressure_fan_min_speed_bar` | Gauge | bar | `target`, `device`, `equip_id` | Minimum speed of the fan |
| `vertiv_ac_pressure_comp_min_output_bar` | Gauge | bar | `target`, `device`, `equip_id` | Compressor minimum output low pressure point |
| `vertiv_ac_pressure_comp_capacity_reduce_bar` | Gauge | bar | `target`, `device`, `equip_id` | Compressor capacity reduces low pressure point |
| `vertiv_ac_pressure_comp_capacity_increase_bar` | Gauge | bar | `target`, `device`, `equip_id` | Compressor capacity increases low pressure point |
| `vertiv_ac_pressure_comp_max_output_bar` | Gauge | bar | `target`, `device`, `equip_id` | Compressor maximum output low pressure point |
| `vertiv_ac_electrical_voltage_phase_a_volts` | Gauge | V | `target`, `device`, `equip_id` | Phase A voltage |
| `vertiv_ac_electrical_voltage_phase_b_volts` | Gauge | V | `target`, `device`, `equip_id` | B phase voltage |
| `vertiv_ac_electrical_voltage_phase_c_volts` | Gauge | V | `target`, `device`, `equip_id` | C phase voltage |
| `vertiv_ac_electrical_frequency_hz` | Gauge | Hz | `target`, `device`, `equip_id` | Power frequency |
| `vertiv_ac_electrical_overvoltage_alarm_percent` | Gauge | % | `target`, `device`, `equip_id` | Power overrun alarm value |
| `vertiv_ac_electrical_undervoltage_alarm_percent` | Gauge | % | `target`, `device`, `equip_id` | Power undervoltage alarm value |
| `vertiv_ac_compressor_capacity_actual_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor capacity actual value |
| `vertiv_ac_compressor_capacity_output_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor capacity output value |
| `vertiv_ac_compressor_min_runtime_minutes` | Gauge | min | `target`, `device`, `equip_id` | Compressor shortest running time |
| `vertiv_ac_compressor_min_downtime_minutes` | Gauge | min | `target`, `device`, `equip_id` | Compressor shortest downtime |
| `vertiv_ac_compressor_start_demand_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor start demand |
| `vertiv_ac_compressor_stop_demand_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor stops demand |
| `vertiv_ac_compressor_min_capacity_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor minimum capacity |
| `vertiv_ac_compressor_standard_capacity_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor standard capacity |
| `vertiv_ac_compressor_max_capacity_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor maximum capacity |
| `vertiv_ac_compressor_start_capacity_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor start capacity |
| `vertiv_ac_compressor_dehumid_capacity_increase_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor dehumidification capacity increases |
| `vertiv_ac_compressor_max_capacity_runtime_minutes` | Gauge | min | `target`, `device`, `equip_id` | Maximum capacity of the compressor running time |
| `vertiv_ac_compressor_start_time_seconds` | Gauge | s | `target`, `device`, `equip_id` | Compressor start time |
| `vertiv_ac_compressor_output_dead_zone_percent` | Gauge | % | `target`, `device`, `equip_id` | Compressor output dead zone |
| `vertiv_ac_compressor_oil_return_cycle_minutes` | Gauge | min | `target`, `device`, `equip_id` | Oil return cycle |
| `vertiv_ac_compressor_oil_return_runtime_minutes` | Gauge | min | `target`, `device`, `equip_id` | Oil return running time |
| `vertiv_ac_compressor_oil_return_capacity_percent` | Gauge | % | `target`, `device`, `equip_id` | Oil return capacity |
| `vertiv_ac_compressor_temp_proportional_band_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Compressor temperature proportional band |
| `vertiv_ac_compressor_temp_integration_time_seconds` | Gauge | s | `target`, `device`, `equip_id` | Compressor temperature integration time |
| `vertiv_ac_compressor_temp_differential_time_seconds` | Gauge | s | `target`, `device`, `equip_id` | Compressor temperature differential time |
| `vertiv_ac_compressor_running_hours_total` | Gauge | h | `target`, `device`, `equip_id` | Compressor running hours |
| `vertiv_ac_compressor_startstop_records_total` | Gauge | 次 | `target`, `device`, `equip_id` | Number of compressor start and stop records |
| `vertiv_ac_compressor_high_pressure_anomalies_total` | Gauge | 次 | `target`, `device`, `equip_id` | Number of high pressure anomalies recorded |
| `vertiv_ac_fan_speed_percent` | Gauge | % | `target`, `device`, `equip_id` | Fan speed |
| `vertiv_ac_fan_min_speed_percent` | Gauge | % | `target`, `device`, `equip_id` | Fan minimum speed |
| `vertiv_ac_fan_standard_speed_percent` | Gauge | % | `target`, `device`, `equip_id` | Fan standard speed |
| `vertiv_ac_fan_humidification_speed_percent` | Gauge | % | `target`, `device`, `equip_id` | Fan humidification speed |
| `vertiv_ac_fan_analog_output_lower_percent` | Gauge | % | `target`, `device`, `equip_id` | Fan analog output lower limit |
| `vertiv_ac_fan_analog_output_upper_percent` | Gauge | % | `target`, `device`, `equip_id` | Fan analog output upper limit |
| `vertiv_ac_fan_low_speed_step_percent_per_sec` | Gauge | %/s | `target`, `device`, `equip_id` | Fan low speed step |
| `vertiv_ac_fan_high_speed_step_percent_per_sec` | Gauge | %/s | `target`, `device`, `equip_id` | Fan high speed step |
| `vertiv_ac_fan_down_delay_seconds` | Gauge | s | `target`, `device`, `equip_id` | Fan down delay |
| `vertiv_ac_fan_start_delay_seconds` | Gauge | s | `target`, `device`, `equip_id` | Fan start delay |
| `vertiv_ac_fan_downtime_seconds` | Gauge | s | `target`, `device`, `equip_id` | Fan downtime |
| `vertiv_ac_fan_temp_proportional_band_celsius` | Gauge | °C | `target`, `device`, `equip_id` | Fan temperature proportional band |
| `vertiv_ac_fan_temp_integration_time_seconds` | Gauge | s | `target`, `device`, `equip_id` | Fan temperature integration time |
| `vertiv_ac_fan_running_hours_total` | Gauge | h | `target`, `device`, `equip_id` | Fan running hours |
| `vertiv_ac_fan_startstop_records_total` | Gauge | 次 | `target`, `device`, `equip_id` | Number of fan start and stop records |
| `vertiv_ac_eev_opening_degree_percent` | Gauge | % | `target`, `device`, `equip_id` | Expansion valve opening degree |
| `vertiv_ac_eev_time_constant_seconds` | Gauge | s | `target`, `device`, `equip_id` | EEV time constant |
| `vertiv_ac_eev_start_opening_percent` | Gauge | % | `target`, `device`, `equip_id` | EEV start opening degree |
| `vertiv_ac_eev_start_time_seconds` | Gauge | s | `target`, `device`, `equip_id` | EEV start time |
| `vertiv_ac_pump_running_hours_total` | Gauge | h | `target`, `device`, `equip_id` | Pump running hours |
| `vertiv_ac_pump_startstop_records_total` | Gauge | 次 | `target`, `device`, `equip_id` | Number of pump start and stop records |
| `vertiv_ac_humidifier_running_hours_total` | Gauge | h | `target`, `device`, `equip_id` | Humidifier running hours |
| `vertiv_ac_humidifier_startstop_records_total` | Gauge | 次 | `target`, `device`, `equip_id` | Number of humidifier start and stop records |
| `vertiv_ac_electric_heating_running_hours_total` | Gauge | h | `target`, `device`, `equip_id` | Electric heating operation hours |
| `vertiv_ac_electric_heating_startstop_records_total` | Gauge | 次 | `target`, `device`, `equip_id` | Number of electric heating start and stop records |
| `vertiv_ac_dehumidification_runtime_minutes` | Gauge | min | `target`, `device`, `equip_id` | Dehumidification run time |
| `vertiv_ac_filter_maintenance_interval_days` | Gauge | d | `target`, `device`, `equip_id` | Filter maintenance reminder time |
| `vertiv_ac_status_operation` | Gauge | — | `target`, `device`, `equip_id` | Air conditioning operation status |
| `vertiv_ac_status_refrigeration` | Gauge | — | `target`, `device`, `equip_id` | Refrigeration flag |
| `vertiv_ac_status_heating` | Gauge | — | `target`, `device`, `equip_id` | Heating flag |
| `vertiv_ac_status_humidification` | Gauge | — | `target`, `device`, `equip_id` | Humidification mark |
| `vertiv_ac_status_dehumidification` | Gauge | — | `target`, `device`, `equip_id` | Dehumidification mark |
| `vertiv_ac_status_compressor_output` | Gauge | — | `target`, `device`, `equip_id` | Compressor output |
| `vertiv_ac_status_fan_output` | Gauge | — | `target`, `device`, `equip_id` | Fan output |
| `vertiv_ac_status_electric_heating_output` | Gauge | — | `target`, `device`, `equip_id` | Electric heating output |
| `vertiv_ac_status_humidifier_output` | Gauge | — | `target`, `device`, `equip_id` | Humidifier output |
| `vertiv_ac_status_condensate_pump_output` | Gauge | — | `target`, `device`, `equip_id` | Condensate pump output |
| `vertiv_ac_status_liquid_solenoid_valve_output` | Gauge | — | `target`, `device`, `equip_id` | Liquid circuit solenoid valve output |
| `vertiv_ac_status_public_alarm_output` | Gauge | — | `target`, `device`, `equip_id` | Public alarm output |
| `vertiv_ac_status_high_pressure_alarm` | Gauge | — | `target`, `device`, `equip_id` | High pressure alarm |
| `vertiv_ac_status_water_level_switch` | Gauge | — | `target`, `device`, `equip_id` | Water level switch |
| `vertiv_ac_status_remote_switch` | Gauge | — | `target`, `device`, `equip_id` | Remote switch |
| `vertiv_ac_status_fan_1` | Gauge | — | `target`, `device`, `equip_id` | Fan 1 state |
| `vertiv_ac_status_fan_2` | Gauge | — | `target`, `device`, `equip_id` | Fan 2 state |
| `vertiv_ac_status_fan_3` | Gauge | — | `target`, `device`, `equip_id` | Fan 3 state |
| `vertiv_ac_status_fan_4` | Gauge | — | `target`, `device`, `equip_id` | Fan 4 state |
| `vertiv_ac_status_new_alarm_flag` | Gauge | — | `target`, `device`, `equip_id` | New alarm flag |
| `vertiv_ac_status_filter_maintenance` | Gauge | — | `target`, `device`, `equip_id` | Filter maintenance |
| `vertiv_ac_status_communication` | Gauge | — | `target`, `device`, `equip_id` | Communication Status |
| `vertiv_ac_status_dehumidification_enabled` | Gauge | — | `target`, `device`, `equip_id` | Dehumidification function enabled |
| `vertiv_ac_status_humidification_enabled` | Gauge | — | `target`, `device`, `equip_id` | Humidification function enabled |
| `vertiv_ac_status_heating_function` | Gauge | — | `target`, `device`, `equip_id` | Heating function |
| `vertiv_ac_status_condensate_pump` | Gauge | — | `target`, `device`, `equip_id` | Condensate pump |
| `vertiv_ac_status_monitor_shutdown_enable` | Gauge | — | `target`, `device`, `equip_id` | Monitor shutdown enable |
| `vertiv_ac_status_soft_shutdown` | Gauge | — | `target`, `device`, `equip_id` | Soft shutdown status |
| `vertiv_ac_alarm_attr_high_voltage` | Gauge | — | `target`, `device`, `equip_id` | High voltage alarm attribute |
| `vertiv_ac_alarm_attr_low_voltage` | Gauge | — | `target`, `device`, `equip_id` | Low voltage alarm attribute |
| `vertiv_ac_alarm_attr_exhaust_high_temp` | Gauge | — | `target`, `device`, `equip_id` | Exhaust high temperature alarm attribute |
| `vertiv_ac_alarm_attr_exhaust_superheat_low` | Gauge | — | `target`, `device`, `equip_id` | Exhaust superheat low alarm attribute |
| `vertiv_ac_alarm_attr_return_air_temp` | Gauge | — | `target`, `device`, `equip_id` | Return air temperature alarm attribute |
| `vertiv_ac_alarm_attr_airflow_temp` | Gauge | — | `target`, `device`, `equip_id` | Air temperature alarm attribute |
| `vertiv_ac_alarm_attr_return_air_humidity` | Gauge | — | `target`, `device`, `equip_id` | Return air humidity alarm attribute |
| `vertiv_ac_alarm_attr_return_air_low_humidity` | Gauge | — | `target`, `device`, `equip_id` | Return air low humidity alarm attribute |
| `vertiv_ac_alarm_attr_high_voltage_lock` | Gauge | — | `target`, `device`, `equip_id` | High voltage lock alarm attribute |
| `vertiv_ac_alarm_attr_low_voltage_lock` | Gauge | — | `target`, `device`, `equip_id` | Low-voltage lock alarm attribute |
| `vertiv_ac_alarm_attr_power_loss` | Gauge | — | `target`, `device`, `equip_id` | Power loss alarm attribute |
| `vertiv_ac_alarm_attr_power_overvoltage` | Gauge | — | `target`, `device`, `equip_id` | Power overvoltage alarm attribute |
| `vertiv_ac_alarm_attr_power_undervoltage` | Gauge | — | `target`, `device`, `equip_id` | Power undervoltage alarm attribute |
| `vertiv_ac_alarm_attr_floor_overflow` | Gauge | — | `target`, `device`, `equip_id` | Floor overflow alarm attribute |
| `vertiv_ac_alarm_attr_high_water` | Gauge | — | `target`, `device`, `equip_id` | High water alarm attribute |
| `vertiv_ac_alarm_attr_filter_plugging` | Gauge | — | `target`, `device`, `equip_id` | Filter plugging alarm attribute |
| `vertiv_ac_alarm_attr_airflow_loss` | Gauge | — | `target`, `device`, `equip_id` | Airflow loss alarm attribute |
| `vertiv_ac_alarm_attr_remote_shutdown` | Gauge | — | `target`, `device`, `equip_id` | Remote shutdown alarm attribute |
| `vertiv_ac_alarm_attr_return_air_temp_sensor_fault` | Gauge | — | `target`, `device`, `equip_id` | Return air temperature sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_return_air_humidity_sensor_fault` | Gauge | — | `target`, `device`, `equip_id` | Return air humidity sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_exhaust_temp_sensor_fault` | Gauge | — | `target`, `device`, `equip_id` | Exhaust temperature sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_fan_failure` | Gauge | — | `target`, `device`, `equip_id` | Fan failure alarm attribute |
| `vertiv_ac_alarm_attr_eev_comm_fault` | Gauge | — | `target`, `device`, `equip_id` | EEV communication fault alarm attribute |
| `vertiv_ac_alarm_attr_insufficient_refrigerant` | Gauge | — | `target`, `device`, `equip_id` | Insufficient refrigerant alarm attribute |
| `vertiv_ac_alarm_attr_inspiratory_temp_sensor_fault` | Gauge | — | `target`, `device`, `equip_id` | Inhalation temperature sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_compressor_drive_comm_fault` | Gauge | — | `target`, `device`, `equip_id` | Compressor Drive Communication Fault Alarm Attribute |
| `vertiv_ac_alarm_attr_compressor_drive_failure` | Gauge | — | `target`, `device`, `equip_id` | Compressor drive failure failure alarm |
| `vertiv_ac_alarm_attr_compressor_radiator_over_temp` | Gauge | — | `target`, `device`, `equip_id` | Compressor radiator over temperature alarm |
| `vertiv_ac_alarm_attr_compressor_overcurrent` | Gauge | — | `target`, `device`, `equip_id` | Compressor overcurrent alarm |
| `vertiv_ac_alarm_attr_compressor_phase_failure` | Gauge | — | `target`, `device`, `equip_id` | Compressor phase failure protection alarm |
| `vertiv_ac_alarm_attr_busbar_voltage_exception` | Gauge | — | `target`, `device`, `equip_id` | Busbar voltage exception alarm |
| `vertiv_ac_alarm_attr_humidifier_fault` | Gauge | — | `target`, `device`, `equip_id` | Humidifier fault alarm attribute |
| `vertiv_ac_system_alarm_active_count` | Gauge | 个 | `target`, `device`, `equip_id` | Number of alarm states |
| `vertiv_ac_system_alarm_history_count` | Gauge | 个 | `target`, `device`, `equip_id` | Number of alarm history |
| `vertiv_ac_system_software_version_high` | Gauge | — | `target`, `device`, `equip_id` | Software version is high |
| `vertiv_ac_system_software_version_low` | Gauge | — | `target`, `device`, `equip_id` | The software version is low |
| `vertiv_ac_system_monitor_baud_rate` | Gauge | baud | `target`, `device`, `equip_id` | Monitor baud rate |
| `vertiv_ac_system_monitor_address` | Gauge | — | `target`, `device`, `equip_id` | Monitor the address |
| `vertiv_ac_system_unit_count` | Gauge | 个 | `target`, `device`, `equip_id` | Number of units |
| `vertiv_ac_system_main_delay_minutes` | Gauge | min | `target`, `device`, `equip_id` | Main delay |
| `vertiv_ac_system_low_pressure_alarm_delay_seconds` | Gauge | s | `target`, `device`, `equip_id` | Low pressure alarm delay |
| `vertiv_ac_system_short_cycle_alarm_times_per_hour` | Gauge | 次/小时 | `target`, `device`, `equip_id` | Short cycle alarm value |
| `vertiv_ac_system_exhaust_superheat_low_alarm_delay_sec` | Gauge | s | `target`, `device`, `equip_id` | Exhaust superheat low alarm delay |
| `vertiv_ac_signal_value` | Gauge | 动态（设备返回，当前未导出单位） | `target`, `device`, `equip_id`, `signal_name`, `occurrence` | 设备返回的原始 AC 信号值；`signal_name` 和 `occurrence` 为动态 Label，单位未被导出。 |

## Run

Use the example config as a starting point:

```bash
cp .config.yaml config.yaml
```

Run tests:

```bash
mkdir -p .gocache .gomodcache
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache GOPROXY=https://proxy.golang.org,direct go test ./...
```

Build:

```bash
mkdir -p .gocache .gomodcache
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache GOPROXY=https://proxy.golang.org,direct go build -o vertiv_exporter ./cmd/vertiv_exporter
```

Start the exporter:

```bash
mkdir -p .gocache .gomodcache
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache GOPROXY=https://proxy.golang.org,direct go run ./cmd/vertiv_exporter -config.file config.yaml
```

Use `--version` to print the version, commit, and build date embedded at build time. The process handles `SIGINT` and `SIGTERM`, cancels in-flight scrapes, shuts down the HTTP server, and stops keepalive goroutines gracefully.

Then open:

- `http://127.0.0.1:9101/`
- `http://127.0.0.1:9101/metrics`

## Binary Releases and Versioning

Official versions follow semantic versioning and use a Git tag such as `v1.2.3` as the single version source. Prerelease tags such as `v1.2.3-rc.1` are supported. Build metadata suffixes such as `+build.1` are rejected so GitHub Release and OCI image tags remain identical. The release binary reports the tag version without the leading `v`, the full commit SHA, and the UTC commit time:

```text
vertiv_exporter version=1.2.3 commit=<full-commit-sha> build_date=<UTC-RFC3339>
```

An ordinary local `go build` keeps the development defaults `version=dev`, `commit=unknown`, and `build_date=unknown`. Every official release contains:

- `vertiv_exporter-<version>.linux-amd64.tar.gz`
- `vertiv_exporter-<version>.linux-arm64.tar.gz`
- `sha256sums.txt`

Download and verify a release on Linux:

```bash
RELEASE=v1.2.3
ARCH=amd64 # use arm64 on 64-bit ARM systems
PACKAGE="vertiv_exporter-${RELEASE#v}.linux-${ARCH}"

curl -fLO "https://github.com/airsmon/vertiv_exporter/releases/download/${RELEASE}/${PACKAGE}.tar.gz"
curl -fLO "https://github.com/airsmon/vertiv_exporter/releases/download/${RELEASE}/sha256sums.txt"
grep -F "  ${PACKAGE}.tar.gz" sha256sums.txt | sha256sum --check -
tar -xzf "${PACKAGE}.tar.gz"
cd "${PACKAGE}"
./vertiv_exporter --version
```

Each archive contains the binary, `.config.yaml`, `vertiv_exporter.service`, and this README. To run the binary directly:

```bash
cp .config.yaml config.yaml
# Edit config.yaml before starting.
./vertiv_exporter --config.file=config.yaml
```

## systemd

The supplied unit runs the exporter as an unprivileged `vertiv_exporter` service account, reads `/etc/vertiv_exporter/config.yaml`, and starts `/usr/local/bin/vertiv_exporter`.

From an extracted release archive, create the account and install the files:

```bash
sudo useradd --system --user-group --no-create-home \
  --shell /usr/sbin/nologin vertiv_exporter
sudo install -o root -g root -m 0755 \
  vertiv_exporter /usr/local/bin/vertiv_exporter
sudo install -d -o root -g vertiv_exporter -m 0750 /etc/vertiv_exporter
sudo install -o root -g vertiv_exporter -m 0640 \
  .config.yaml /etc/vertiv_exporter/config.yaml
sudo install -o root -g root -m 0644 \
  vertiv_exporter.service /etc/systemd/system/vertiv_exporter.service
sudoedit /etc/vertiv_exporter/config.yaml
```

If your distribution provides `nologin` at `/sbin/nologin`, use that path in the `useradd` command. systemd 245 or newer applies every hardening directive in the supplied unit; older releases may ignore unsupported directives with a warning.

The configuration can contain credentials, so keep it owned by `root:vertiv_exporter` with mode `0640`. Put a custom `metrics_file` under `/etc/vertiv_exporter`, use an absolute path in the configuration, and apply the same ownership and mode.

Enable the service and inspect its status:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now vertiv_exporter
sudo systemctl status --no-pager vertiv_exporter
sudo journalctl -u vertiv_exporter -f
```

After changing the configuration, restart the service because the exporter does not implement live reload:

```bash
sudo systemctl restart vertiv_exporter
curl -fsS http://127.0.0.1:9101/metrics
```

To upgrade, download and verify the matching release first, then replace only the binary so the existing configuration remains intact:

```bash
sudo systemctl stop vertiv_exporter
sudo install -o root -g root -m 0755 \
  vertiv_exporter /usr/local/bin/vertiv_exporter
/usr/local/bin/vertiv_exporter --version
sudo systemctl start vertiv_exporter
sudo systemctl status --no-pager vertiv_exporter
```

## Kubernetes

The recommended deployment method is the
[`charts/vertiv-exporter`](charts/vertiv-exporter) Helm chart. It provides
secure pod defaults, schema-validated target configuration, external Secret
integration, an optional ServiceMonitor, and built-in PrometheusRule alerts.
The chart never accepts credentials in its generated ConfigMap.

Create the shared credentials, prepare a non-sensitive values file as described
in the [chart README](charts/vertiv-exporter/README.md), and install it:

```bash
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
kubectl -n monitoring create secret generic vertiv-exporter-credentials \
  --from-literal=username='<VERTIV_USERNAME>' \
  --from-literal=password='<VERTIV_PASSWORD>'

helm upgrade --install vertiv-exporter ./charts/vertiv-exporter \
  --namespace monitoring \
  --values vertiv-values.yaml
```

Use `config.existingSecret` instead when targets require different credential
pairs. Enable `serviceMonitor.enabled` and `prometheusRule.enabled` only when
Prometheus Operator CRDs are installed.

Verify the exporter:

```bash
kubectl -n monitoring rollout status deployment/vertiv-exporter
kubectl -n monitoring port-forward service/vertiv-exporter 9101:9101
```

The metrics endpoint is then available at `http://127.0.0.1:9101/metrics`.
The probes use `/` so they do not initiate Vertiv device scrapes. Do not commit
real credentials to a values file; for production environments, use the
cluster's existing secret-management solution.

Helm upgrades roll the pods when the generated ConfigMap changes. Restart the
Deployment after updating an external configuration Secret or the shared
credential Secret because Helm does not manage those resources:

```bash
kubectl -n monitoring rollout restart deployment/vertiv-exporter
```

## Prometheus Configuration

The Prometheus timeout should be longer than `exporter.scrape_timeout` so the exporter can report target failures cleanly:

```yaml
scrape_configs:
  - job_name: "vertiv"
    scrape_interval: 30s
    scrape_timeout: 15s
    static_configs:
      - targets: ["vertiv-exporter:9101"]
```

Use `127.0.0.1:9101` when Prometheus and the exporter run directly on the same host. The example service name `vertiv-exporter:9101` works only when both containers share a network where that name resolves; otherwise replace it with the exporter address used by your deployment.

Prometheus supplies `job` and `instance` as scrape-target labels. The exporter uses the separate `target` label for the configured Vertiv target name, so the two identities remain unambiguous and do not depend on `honor_labels` behavior.

When upgrading from an exporter version that used `instance` for the Vertiv target name, update PromQL and dashboard selectors to use `target`. Under the common `honor_labels: false` setting, this means replacing exporter-originated `exported_instance` selectors with `target`; keep any ServiceMonitor relabeling that defines the Prometheus scrape-target `instance` unchanged.

## Docker

Local single-platform build:

```bash
docker build \
  --build-arg VERSION=1.0.1 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t vertiv-exporter:1.0.1 .
```

Multi-platform build and push:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=1.0.1 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t your-registry/vertiv-exporter:1.0.1 \
  --push .
```

Run it with a mounted config file:

```bash
docker run --rm -p 9101:9101 \
  --user "$(id -u):$(id -g)" \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  vertiv-exporter:1.0.1
```

The image expects the config file at `/app/config.yaml`. The local example runs with the host UID/GID so a `chmod 600` bind-mounted config remains readable without making it world-readable. Secret or ConfigMap mounts can keep the image's default `nonroot` user when their permissions allow UID `65532` to read the file. If `exporter.metrics_file` is set, mount that override file separately as read-only and use its in-container path in `config.yaml`.

The build context is restricted by `.dockerignore` to the build definition, `go.mod`, `go.sum`, and production Go source files. Local credentials, Git history, dashboards, documentation, tests, caches, and build artifacts are not sent to the Docker builder.

## GitHub CI, Releases, and Container Images

The GitHub Actions CI runs Go formatting, module verification, `go vet`, race-enabled tests, a versioned binary build, Helm lint/render/package checks, and Prometheus rule validation for pull requests, pushes to `main`, and `v*` tags. It then builds and smoke-tests the container on both `linux/amd64` and `linux/arm64`.

After all checks pass, pushes to `main` and `v*` tags publish a multi-platform image to:

```text
ghcr.io/airsmon/vertiv_exporter
```

The workflow uses the repository-provided `GITHUB_TOKEN`; no custom registry secret is required. A `main` push publishes `main` and `sha-*` tags. A stable tag such as `v1.2.3` publishes `v1.2.3`, `1.2.3`, `1.2`, `1`, `latest`, and `sha-*`; prerelease tags do not move `latest`.

After the tag image passes tests and is pushed successfully, the same workflow builds the Linux release archives, verifies their embedded version metadata and checksums, and creates a GitHub Release with generated notes. Create both the release image and binary release by pushing a semantic version tag:

```bash
git tag v1.2.3
git push origin v1.2.3
```

Pull the latest stable image with:

```bash
docker pull ghcr.io/airsmon/vertiv_exporter:latest
```

The first published package may be private depending on the repository or organization defaults. Change its visibility in the package settings when public anonymous pulls are required.

In addition, `govulncheck` runs for relevant pull requests and pushes, once a week, and on demand. Dependabot checks Go modules, GitHub Actions, and the Docker base images monthly.

## Grafana Dashboard

推荐导入 [`grafana/vertiv_grafana_dashboard.json`](grafana/vertiv_grafana_dashboard.json)。该 dashboard 的 UID 为 `vertiv-exporter-overview`，所有查询使用 `${DS_PROMETHEUS}` 数据源变量，并提供 `job`、`target`、`device` 筛选。详细导入步骤和面板指标清单见 [`grafana/README.md`](grafana/README.md)。

## Supported Metric Groups

Every device metric includes `target`, `device`, and `equip_id`. THD and UPS metrics add the labels noted below. Exporter health is reported through:

- `vertiv_exporter_up{target}`: `1` when every configured device for the target was collected, otherwise `0`
- `vertiv_exporter_scrape_duration_seconds`: histogram of complete collector run duration
- `vertiv_exporter_scrape_failures_total`: total device fetch failures

### AC

- Temperature
- Humidity
- Pressure
- Electrical values
- Compressor, fan, EEV, runtime, alarm attributes, system status
- `vertiv_ac_signal_value{signal_name,occurrence}` exposes every named AC signal
  reported by the device, including signals without a built-in field mapping.
  Existing semantic AC metrics remain available for dashboard compatibility.

The raw AC metric uses only the signal name and parsed numeric value from the
device response. Source indexes and sampling timestamps are not exported.
`occurrence` is normally `1` and increases only when one device reports the
same signal name more than once.

### THD

- `vertiv_thd_temperature_celsius{rack,aisle,position}`
- `vertiv_thd_humidity_percent{rack,aisle}`
- `vertiv_thd_door_status{rack,aisle}`
- `vertiv_thd_comm_status{rack}`
- `vertiv_thd_rack_id{rack}`
- `vertiv_thd_high_temp_alarm_rack_count`

THD temperature and humidity values are rounded to 2 decimal places before export.

### UPS

- Input phase and line voltage/current/frequency/power factor
- Bypass phase and line voltage/frequency
- Output phase voltage/current/frequency/power factor
- Output active power, apparent power, load percent
- Battery, runtime, energy, environment, and status metrics

## Notes and Optimizations

- Explicit `device.type` is supported and recommended so device intent remains clear when names vary.
- THD metrics are label-merged instead of exploded into many independent metric names.
- UPS metrics are field-id based and label-driven, which is more stable than parsing English field names.
- Parse errors include a response preview to make field troubleshooting faster.

## Known Behavior

- Prometheus stores numeric values, not display formatting strings.
- For THD metrics, values are rounded to 2 decimal places in code, but `/metrics` may still render `40.4` instead of `40.40`. That is normal Prometheus float formatting behavior.
