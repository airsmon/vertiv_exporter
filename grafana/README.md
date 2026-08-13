# Vertiv Exporter Grafana Dashboard

`vertiv_grafana_dashboard.json` 是可直接导入 Grafana 的标准 dashboard JSON，默认兼容 Grafana 10 及以上版本。

## 导入

1. 在 Grafana 中打开 **Dashboards → New → Import**。
2. 上传 [`vertiv_grafana_dashboard.json`](vertiv_grafana_dashboard.json)。
3. 在导入页为 `DS_PROMETHEUS` 选择实际的 Prometheus 数据源。
4. 导入后通过顶部的 `job`、`target` 和 `device` 变量切换采集任务、Vertiv 控制器和 AC 设备。仓库内 Dashboard 的默认控制器为 `SH-SP-06-7`，变量选项仍由 Prometheus 动态生成。

Dashboard 不包含硬编码的数据源 UID。所有面板均使用 `${DS_PROMETHEUS}`；控制器由 Exporter 显式导出的 `target` Label 筛选。Prometheus 的 `instance` 仅表示抓取端点，不参与控制器选择，因此 Pod 重建不会在变量中产生新的控制器。

## 阈值与颜色原则

Dashboard 只为 Vertiv 官方资料能够证明、且跨目标仍有明确语义的值设置健康色。设备型号、控制器 setpoint、电池串数量或站点策略不同会改变有效范围，因此不会为所有数值机械套用一组红黄绿：

- 设备通信、供电、门、告警等离散状态使用明确的 Value Mapping。
- 压力、温湿度只在厂商公布了范围时使用数值阈值；厂商仅说明“超过控制器设定值”的项目保持中性色，并以设备自身告警为准。
- 容量、风机转速、功率、电流等负载相关值是运行输出，不因数值较高直接判故障。
- 累计能耗、累计运行时间、启停次数、放电次数和历史告警数都是生命周期累计量，不设置“超过某值永久红色”的健康阈值。

当前阈值依据：

- [Vertiv Liebert CRV+ 用户手册](https://www.vertiv.com/4a7ad2/globalassets/products/thermal-management/in-row-cooling/liebert-crv2/liebert-crv-series-air-conditioner-user-manual.pdf)：吸气压力 4.5–11 bar、R410A 高压开关 37 bar，以及送风/回风温湿度默认告警与可调范围。
- [Vertiv SmartRow 2 用户指南](https://www.vertiv.com/4a5dbe/globalassets/shared/vertiv-smartrow-2-user-guide-sl-71214.pdf)：RDU-THD 冷通道 18–27°C、热通道 18–40°C、湿度 8–95%，门打开为 Maintenance 状态。
- [Vertiv Liebert ITA2 30kVA UPS 用户手册](https://www.vertiv.com/4ace67/globalassets/products/critical-power/uninterruptible-power-supplies-ups/liebert-ita2-30kva-ups-user-manual_v1.2.pdf)：UPS 环境、电气与负载运行范围。

Vertiv 公开 CRV/CRD 手册没有提供跨型号通用的高排气温度数值，只说明超过控制器 set value 会告警。因此 Dashboard 不再使用原先 `<60°C` 或 `≥85°C` 的猜测阈值。CRV+ 排气压力只公开 37 bar 高压开关动作点、没有排气压力下限，所以低于 20 bar 也不会被判为故障。

UPS 输入/输出能耗是控制器生命周期累计 `kWh` Counter。Dashboard 的专用面板除以 1000 显示为 `MWh`，例如 `268528 kWh = 268.528 MWh`；这只是显示换算，不会修改原始指标，也不表示当前时段消耗了 268 MWh。

## Label 迁移兼容

当前 Dashboard 同时兼容新的 `target` Label 和保留期内的旧数据。旧版 Exporter 将控制器名称导出为 `instance`；在 Prometheus 默认的 `honor_labels: false` 配置下，这部分历史数据存储为 `exported_instance`。兼容分支仅选择缺少原生 `target` 的旧样本，再通过 `label_replace` 将 `exported_instance` 规范化为 `target`：

```promql
max without (instance, pod, exported_instance) (
  metric{job=~"$job", target=~"$target", target!=""}
  or
  label_replace(
    metric{
      job=~"$job",
      target="",
      exported_instance=~"$target",
      exported_instance!=""
    },
    "target", "$1", "exported_instance", "(.+)"
  )
)
```

`target!=""` 和旧分支的 `target=""` 保证同一份新样本不会被读取两次；外层聚合去除旧 Pod IP、Pod 名和迁移 Label，使同一控制器在迁移前后的数据保持为一条时序。多指标查询在聚合前将 `__name__` 复制到 `metric`，避免不同指标被合并，图例相应使用 `{{metric}}`。

Exporter 自身的 `vertiv_exporter_scrape_failures_total` 仍按 Counter 处理：先对每个抓取时序执行 `increase`，再按 `job` 求和。UPS 累计 Counter 在表格中展示当前累计值，因此与其他当前值一样仅使用兼容聚合，不计算速率。AC 名称以 `_total` 结尾的字段目前由 Exporter 作为 Gauge 导出，也不使用 `rate` 或 `increase`。

当前 10 天旧数据超过 Prometheus 保留期后（建议额外保留一天安全余量），可以删除 `exported_instance` 兼容分支；`target` 变量和查询接口无需再次变更。

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
| UPS Battery & Runtime Inventory | `vertiv_ups_battery_capacity_percent`, `vertiv_ups_battery_backup_time_minutes`, `vertiv_ups_battery_discharging_time_seconds`, `vertiv_ups_battery_discharge_count_total`, `vertiv_ups_ambient_temperature_celsius`, `vertiv_ups_running_time_days`, `vertiv_ups_parallel_machine_count` |
| UPS Lifetime Energy (MWh) | `vertiv_ups_input_energy_kwh_total`, `vertiv_ups_output_energy_kwh_total`（查询时除以 1000，仅用于 MWh 展示） |
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
