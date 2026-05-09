# Vertiv AC Prometheus 指标完整列表

> 数据来源：`ENP_AC_SRVII[COM]` 设备，共解析 **~200** 个数据点  
> Labels（所有指标共有）：`instance="dc-rack-01"` `device="AC1"` `equip_id="23"`  
> 命名规范：`vertiv_ac_<category>_<name>_<unit>`

---

## 一、温度类 `vertiv_ac_temperature_*_celsius` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 |
|--------|---------|-----------|--------|
| `vertiv_ac_temperature_return_air_celsius` | 2 | Return air temperature measurement | 28.6 |
| `vertiv_ac_temperature_supply_air_1_celsius` | 157 | Air supply temperature 1 measured value | 19.4 |
| `vertiv_ac_temperature_supply_air_2_celsius` | 263 | Air supply temperature 2 measured value | 18.3 |
| `vertiv_ac_temperature_supply_air_mean_celsius` | 325 | Mean temperature measurement of Supply Air | 18.8 |
| `vertiv_ac_temperature_supply_air_setpoint_celsius` | 323 | Supply Air temperature setting | 20.0 |
| `vertiv_ac_temperature_airflow_celsius` | 9 | Air temperature measurement | 30.2 |
| `vertiv_ac_temperature_exhaust_celsius` | 10 | Exhaust temperature measurement | 55.0 |
| `vertiv_ac_temperature_inspiratory_celsius` | 324 | Inspiratory temperature measurement | 21.2 |
| `vertiv_ac_temperature_inspiratory_evaporation_celsius` | 34 | Inspiratory evaporation temperature | 12.2 |
| `vertiv_ac_temperature_exhaust_condensing_celsius` | 35 | Exhaust condensing temperature | 34.4 |
| `vertiv_ac_temperature_inspiratory_superheat_celsius` | 36 | Inspiratory superheat | 9.0 |
| `vertiv_ac_temperature_exhaust_superheat_celsius` | 37 | Exhaust superheat | 20.6 |
| `vertiv_ac_temperature_return_air_setpoint_celsius` | 114 | Return air temperature setting | 30.0 |
| `vertiv_ac_temperature_remote_setpoint_celsius` | 113 | Remote temperature setting | 20.0 |
| `vertiv_ac_temperature_airflow_alarm_value_celsius` | 141 | Air temperature alarm value | 30.0 |
| `vertiv_ac_temperature_low_alarm_value_celsius` | 142 | Low temperature alarm value | 8.0 |
| `vertiv_ac_temperature_return_air_alarm_value_celsius` | 143 | Return air temperature alarm value | 35.0 |
| `vertiv_ac_temperature_airflow_loss_alarm_celsius` | 146 | Airflow loss temperature alarm value | 16.0 |
| `vertiv_ac_temperature_dead_zone_celsius` | 111 | Temperature dead zone | 0.5 |
| `vertiv_ac_temperature_dehumid_stop_diff_celsius` | 127 | Dehumidification stop temperature difference | -5.0 |
| `vertiv_ac_temperature_eev_superheat_setpoint_celsius` | 183 | EEV superheat setting | 6.0 |
| `vertiv_ac_temperature_eev_close_superheat_celsius` | 186 | The EEV valve closes the superheat | 6.0 |
| `vertiv_ac_temperature_remote_1_celsius` | 130 | Remote temperature 1 measurements | 0.0 |
| `vertiv_ac_temperature_remote_2_celsius` | 131 | Remote temperature 2 measurements | 0.0 |
| `vertiv_ac_temperature_remote_3_celsius` | 132 | Remote temperature 3 measurements | 0.0 |
| `vertiv_ac_temperature_remote_avg_celsius` | 140 | Remote average temperature | 0.0 |
| `vertiv_ac_temperature_supply_air_1_correction_celsius` | 15 | Air supply temperature 1 correction value | 0.0 |
| `vertiv_ac_temperature_supply_air_2_correction_celsius` | 16 | Air supply temperature 2 correction value | 0.0 |
| `vertiv_ac_temperature_return_air_correction_celsius` | 17 | Return air temperature correction value | 0.0 |
| `vertiv_ac_temperature_airflow_correction_celsius` | 18 | Airflow temperature correction value | 0.0 |
| `vertiv_ac_temperature_exhaust_correction_celsius` | 19 | Exhaust temperature correction value | 0.0 |
| `vertiv_ac_temperature_inspiratory_correction_celsius` | 20 | Inspiratory temperature correction value | 0.0 |

---

## 二、湿度类 `vertiv_ac_humidity_*` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_humidity_return_air_percent` | 3 | Return air humidity measurement | 30.4 | % |
| `vertiv_ac_humidity_return_air_setpoint_percent` | 4 | Humidity setting | 50.0 | % |
| `vertiv_ac_humidity_ratio_percent` | 65 | Humidity ratio | 5.0 | % |
| `vertiv_ac_humidity_supply_air_theoretical_percent` | 38 | Theoretical air supply humidity | 51.7 | % |
| `vertiv_ac_humidity_supply_air_current_percent` | 39 | Current air supply humidity | 55.7 | % |
| `vertiv_ac_humidity_dead_zone_percent` | 112 | Humidity dead zone | 3.0 | % |
| `vertiv_ac_humidity_return_air_high_alarm_percent` | 144 | Return wind high humidity alarm value | 95.0 | % |
| `vertiv_ac_humidity_return_air_low_alarm_percent` | 145 | Return air low warning value | 8.0 | % |
| `vertiv_ac_humidity_return_air_correction_percent` | 21 | Return air humidity correction value | 0.0 | % |

---

## 三、压力类 `vertiv_ac_pressure_*_bar` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 |
|--------|---------|-----------|--------|
| `vertiv_ac_pressure_exhaust_bar` | 326 | Exhaust pressure measurement | 20.1 |
| `vertiv_ac_pressure_inspiratory_bar` | 14 | Inspiratory pressure measurement | 10.6 |
| `vertiv_ac_pressure_exhaust_correction_bar` | 22 | Exhaust pressure correction value | 0.0 |
| `vertiv_ac_pressure_inspiratory_correction_bar` | 23 | Inspiratory pressure correction value | 0.0 |
| `vertiv_ac_pressure_eev_mop_limit_bar` | 184 | EEV MOP pressure limit | 11.0 |
| `vertiv_ac_pressure_fan_max_speed_low_point_bar` | 204 | Fan maximum speed low pressure point | 5.8 |
| `vertiv_ac_pressure_fan_speed_low_point_bar` | 205 | Fan speed low point | 7.5 |
| `vertiv_ac_pressure_fan_down_low_point_bar` | 206 | Fan down the low pressure point | 11.5 |
| `vertiv_ac_pressure_fan_min_speed_bar` | 207 | Minimum speed of the fan | 13.5 |
| `vertiv_ac_pressure_comp_min_output_bar` | 208 | Compressor minimum output low pressure point | 5.8 |
| `vertiv_ac_pressure_comp_capacity_reduce_bar` | 209 | Compressor capacity reduces low pressure point | 7.5 |
| `vertiv_ac_pressure_comp_capacity_increase_bar` | 210 | Compressor capacity increases low pressure point | 11.5 |
| `vertiv_ac_pressure_comp_max_output_bar` | 211 | Compressor maximum output low pressure point | 13.5 |

---

## 四、电气类 `vertiv_ac_electrical_*` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_electrical_voltage_phase_a_volts` | 26 | Phase A voltage | 213.8 | V |
| `vertiv_ac_electrical_voltage_phase_b_volts` | 27 | B phase voltage | 214.9 | V |
| `vertiv_ac_electrical_voltage_phase_c_volts` | 28 | C phase voltage | 214.9 | V |
| `vertiv_ac_electrical_frequency_hz` | 29 | Power frequency | 49.9 | HZ |
| `vertiv_ac_electrical_overvoltage_alarm_percent` | 223 | Power overrun alarm value | 12.0 | % |
| `vertiv_ac_electrical_undervoltage_alarm_percent` | 224 | Power undervoltage alarm value | -14.0 | % |

---

## 五、压缩机类 `vertiv_ac_compressor_*` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_compressor_capacity_actual_percent` | 115 | Compressor capacity actual value | 31.0 | % |
| `vertiv_ac_compressor_capacity_output_percent` | 116 | Compressor capacity output value | 31.0 | % |
| `vertiv_ac_compressor_min_runtime_minutes` | 30 | Compressor shortest running time | 15.0 | Min |
| `vertiv_ac_compressor_min_downtime_minutes` | 31 | Compressor shortest downtime | 2.0 | Min |
| `vertiv_ac_compressor_start_demand_percent` | 158 | Compressor start demand | 50.0 | % |
| `vertiv_ac_compressor_stop_demand_percent` | 159 | Compressor stops demand | -150.0 | % |
| `vertiv_ac_compressor_min_capacity_percent` | 160 | Compressor minimum capacity | 30.0 | % |
| `vertiv_ac_compressor_standard_capacity_percent` | 316 | Compressor standard capacity | 100.0 | % |
| `vertiv_ac_compressor_max_capacity_percent` | 162 | Compressor maximum capacity | 125.0 | % |
| `vertiv_ac_compressor_start_capacity_percent` | 164 | Compressor start capacity | 40.0 | % |
| `vertiv_ac_compressor_dehumid_capacity_increase_percent` | 165 | Compressor dehumidification capacity increases | 15.0 | % |
| `vertiv_ac_compressor_max_capacity_runtime_minutes` | 166 | Maximum capacity of the compressor running time | 120.0 | Min |
| `vertiv_ac_compressor_start_time_seconds` | 167 | Compressor start time | 180.0 | Sec |
| `vertiv_ac_compressor_output_dead_zone_percent` | 168 | Compressor output dead zone | 2.8 | % |
| `vertiv_ac_compressor_oil_return_cycle_minutes` | 169 | Oil return cycle | 240.0 | Min |
| `vertiv_ac_compressor_oil_return_runtime_minutes` | 170 | Oil return running time | 5.0 | Min |
| `vertiv_ac_compressor_oil_return_capacity_percent` | 171 | Oil return capacity | 60.0 | % |
| `vertiv_ac_compressor_temp_proportional_band_celsius` | 151 | Compressor temperature proportional band | 5.0 | ℃ |
| `vertiv_ac_compressor_temp_integration_time_seconds` | 152 | Compressor temperature integration time | 300.0 | Sec |
| `vertiv_ac_compressor_temp_differential_time_seconds` | 153 | Compressor temperature differential time | 0.0 | Sec |
| `vertiv_ac_compressor_running_hours_total` | 48 | Compressor running hours | 0.0 | Hour |
| `vertiv_ac_compressor_startstop_records_total` | 43 | Number of compressor start and stop records | 50.0 | - |
| `vertiv_ac_compressor_high_pressure_anomalies_total` | 128 | Number of high pressure anomalies recorded | 14.0 | - |

---

## 六、风机类 `vertiv_ac_fan_*` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_fan_speed_percent` | 117 | Fan speed | 40.0 | % |
| `vertiv_ac_fan_min_speed_percent` | 172 | Fan minimum speed | 40.0 | % |
| `vertiv_ac_fan_standard_speed_percent` | 173 | Fan standard speed | 75.0 | % |
| `vertiv_ac_fan_humidification_speed_percent` | 176 | Fan humidification speed | 75.0 | % |
| `vertiv_ac_fan_analog_output_lower_percent` | 177 | Fan analog output lower limit | 30.0 | % |
| `vertiv_ac_fan_analog_output_upper_percent` | 178 | Fan analog output upper limit | 100.0 | % |
| `vertiv_ac_fan_low_speed_step_percent_per_sec` | 179 | Fan low speed step | 0.1 | %/s |
| `vertiv_ac_fan_high_speed_step_percent_per_sec` | 180 | Fan high speed step | 1.0 | %/s |
| `vertiv_ac_fan_down_delay_seconds` | 181 | Fan down delay | 5.0 | Sec |
| `vertiv_ac_fan_start_delay_seconds` | 32 | Fan start delay | 10.0 | Sec |
| `vertiv_ac_fan_downtime_seconds` | 33 | Fan downtime | 30.0 | Sec |
| `vertiv_ac_fan_temp_proportional_band_celsius` | 154 | Fan temperature proportional band | 13.0 | ℃ |
| `vertiv_ac_fan_temp_integration_time_seconds` | 155 | Fan temperature integration time | 240.0 | Sec |
| `vertiv_ac_fan_running_hours_total` | 50 | Fan running hours | 0.0 | Hour |
| `vertiv_ac_fan_startstop_records_total` | 45 | Number of fan start and stop records | 50.0 | - |

---

## 七、膨胀阀 (EEV) 类 `vertiv_ac_eev_*` (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_eev_opening_degree_percent` | 118 | Expansion valve opening degree | 16.0 | % |
| `vertiv_ac_eev_time_constant_seconds` | 182 | EEV time constant | 60.0 | Sec |
| `vertiv_ac_eev_start_opening_percent` | 185 | EEV start opening degree | 65.0 | % |
| `vertiv_ac_eev_start_time_seconds` | 203 | EEV start time | 60.0 | Sec |

---

## 八、泵与加湿器类 (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_pump_running_hours_total` | 49 | Pump running hours | 1026.0 | Hour |
| `vertiv_ac_pump_startstop_records_total` | 44 | Number of pump start and stop records | 50.0 | - |
| `vertiv_ac_humidifier_running_hours_total` | 52 | Humidifier running hours | 0.0 | Hour |
| `vertiv_ac_humidifier_startstop_records_total` | 47 | Number of humidifier start and stop records | 0.0 | - |
| `vertiv_ac_electric_heating_running_hours_total` | 51 | Electric heating operation hours | 0.0 | Hour |
| `vertiv_ac_electric_heating_startstop_records_total` | 46 | Number of electric heating start and stop records | 1.0 | - |
| `vertiv_ac_dehumidification_runtime_minutes` | 126 | Dehumidification run time | 15.0 | Min |
| `vertiv_ac_filter_maintenance_interval_days` | 123 | Filter maintenance reminder time | 90.0 | Day |

---

## 九、告警与运行状态类 `vertiv_ac_status_*` (Gauge, 0/1)

> 解析规则：枚举值括号内数字直接作为 Gauge 值（如 `Running[0]` → 0，`TurnON[1]` → 1）

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 说明 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_status_operation` | 5 | Air conditioning operation status | Running[0] | 0=Running |
| `vertiv_ac_status_refrigeration` | 8 | Refrigeration flag | In[0] | 1=In |
| `vertiv_ac_status_heating` | 11 | Heating flag | NotIn[1] | 1=NotIn |
| `vertiv_ac_status_humidification` | 12 | Humidification mark | NotIn[1] | - |
| `vertiv_ac_status_dehumidification` | 13 | Dehumidification mark | NotIn[1] | - |
| `vertiv_ac_status_compressor_output` | 261 | Compressor output | OutputMove[1] | 1=运行 |
| `vertiv_ac_status_fan_output` | 7 | Fan output | OutputMove[0] | - |
| `vertiv_ac_status_electric_heating_output` | 262 | Electric heating output | OutputClose[0] | - |
| `vertiv_ac_status_humidifier_output` | 264 | Humidifier output | OutputClose[0] | - |
| `vertiv_ac_status_condensate_pump_output` | 260 | Condensate pump output | OutputClose[0] | - |
| `vertiv_ac_status_liquid_solenoid_valve_output` | 266 | Liquid circuit solenoid valve output | OutputMove[1] | - |
| `vertiv_ac_status_public_alarm_output` | 265 | Public alarm output | OutputMove[1] | 1=告警激活 |
| `vertiv_ac_status_high_pressure_alarm` | 267 | High pressure alarm | OutputClose[0] | - |
| `vertiv_ac_status_water_level_switch` | 258 | Water level switch | Close[0] | - |
| `vertiv_ac_status_remote_switch` | 259 | Remote switch | Close[0] | - |
| `vertiv_ac_status_fan_1` | 161 | Fan 1 state | Normal[0] | 0=Normal |
| `vertiv_ac_status_fan_2` | 317 | Fan 2 state | Normal[0] | - |
| `vertiv_ac_status_fan_3` | 318 | Fan 3 state | Normal[0] | - |
| `vertiv_ac_status_fan_4` | 319 | Fan 4 state | Normal[0] | - |
| `vertiv_ac_status_new_alarm_flag` | 311 | New alarm flag | GenerateNewAlarm[1] | 1=有新告警 |
| `vertiv_ac_status_filter_maintenance` | 315 | Filter maintenance | No[0] | 1=需维护 |
| `vertiv_ac_status_communication` | 91 | Communication Status | Normal[0] | 0=Normal |
| `vertiv_ac_status_dehumidification_enabled` | 306 | Dehumidification function enabled | Enable[1] | - |
| `vertiv_ac_status_humidification_enabled` | 305 | Humidification function enabled | Disable[0] | - |
| `vertiv_ac_status_heating_function` | 101 | Heating function | Disable[0] | - |
| `vertiv_ac_status_condensate_pump` | 103 | Condensate pump | Have[1] | - |
| `vertiv_ac_status_monitor_shutdown_enable` | 309 | Monitor shutdown enable | Enable[1] | - |
| `vertiv_ac_status_soft_shutdown` | 310 | Soft shutdown status | No[0] | - |

---

## 十、告警属性类 `vertiv_ac_alarm_attr_*` (Gauge, 0/1)

> 这类字段描述每类告警是否启用（`TurnON[1]`=启用，`Stop[0]`=禁用）

| 指标名 | 字段 ID | 原始字段名 |
|--------|---------|-----------|
| `vertiv_ac_alarm_attr_high_voltage` | 66 | High voltage alarm attribute |
| `vertiv_ac_alarm_attr_low_voltage` | 67 | Low voltage alarm attribute |
| `vertiv_ac_alarm_attr_exhaust_high_temp` | 69 | Exhaust high temperature alarm attribute |
| `vertiv_ac_alarm_attr_exhaust_superheat_low` | 70 | Exhaust superheat low alarm attribute |
| `vertiv_ac_alarm_attr_return_air_temp` | 71 | Return air temperature alarm attribute |
| `vertiv_ac_alarm_attr_airflow_temp` | 72 | Air temperature alarm attribute |
| `vertiv_ac_alarm_attr_return_air_humidity` | 74 | Return air humidity alarm attribute |
| `vertiv_ac_alarm_attr_return_air_low_humidity` | 75 | Return air low humidity alarm attribute |
| `vertiv_ac_alarm_attr_high_voltage_lock` | 76 | High voltage lock alarm attribute |
| `vertiv_ac_alarm_attr_low_voltage_lock` | 77 | Low-voltage lock alarm attribute |
| `vertiv_ac_alarm_attr_power_loss` | 80 | Power loss alarm attribute |
| `vertiv_ac_alarm_attr_power_overvoltage` | 81 | Power overvoltage alarm attribute |
| `vertiv_ac_alarm_attr_power_undervoltage` | 82 | Power undervoltage alarm attribute |
| `vertiv_ac_alarm_attr_floor_overflow` | 84 | Floor overflow alarm attribute |
| `vertiv_ac_alarm_attr_high_water` | 85 | High water alarm attribute |
| `vertiv_ac_alarm_attr_filter_plugging` | 86 | Filter plugging alarm attribute |
| `vertiv_ac_alarm_attr_airflow_loss` | 88 | Airflow loss alarm attribute |
| `vertiv_ac_alarm_attr_remote_shutdown` | 90 | Remote shutdown alarm attribute |
| `vertiv_ac_alarm_attr_return_air_temp_sensor_fault` | 93 | Return air temperature sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_return_air_humidity_sensor_fault` | 94 | Return air humidity sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_exhaust_temp_sensor_fault` | 188 | Exhaust temperature sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_fan_failure` | 189 | Fan failure alarm attribute |
| `vertiv_ac_alarm_attr_eev_comm_fault` | 190 | EEV communication fault alarm attribute |
| `vertiv_ac_alarm_attr_insufficient_refrigerant` | 192 | Insufficient refrigerant alarm attribute |
| `vertiv_ac_alarm_attr_inspiratory_temp_sensor_fault` | 193 | Inhalation temperature sensor fault alarm attribute |
| `vertiv_ac_alarm_attr_compressor_drive_comm_fault` | 195 | Compressor Drive Communication Fault Alarm Attribute |
| `vertiv_ac_alarm_attr_compressor_drive_failure` | 196 | Compressor drive failure failure alarm |
| `vertiv_ac_alarm_attr_compressor_radiator_over_temp` | 197 | Compressor radiator over temperature alarm |
| `vertiv_ac_alarm_attr_compressor_overcurrent` | 198 | Compressor overcurrent alarm |
| `vertiv_ac_alarm_attr_compressor_phase_failure` | 199 | Compressor phase failure protection alarm |
| `vertiv_ac_alarm_attr_busbar_voltage_exception` | 200 | Busbar voltage exception alarm |
| `vertiv_ac_alarm_attr_humidifier_fault` | 201 | Humidifier fault alarm attribute |

---

## 十一、统计与系统类 (Gauge)

| 指标名 | 字段 ID | 原始字段名 | 示例值 | 单位 |
|--------|---------|-----------|--------|------|
| `vertiv_ac_system_alarm_active_count` | 41 | Number of alarm states | 1.0 | - |
| `vertiv_ac_system_alarm_history_count` | 42 | Number of alarm history | 153.0 | - |
| `vertiv_ac_system_software_version_high` | 55 | Software version is high | 154.0 | - |
| `vertiv_ac_system_software_version_low` | 56 | The software version is low | 0.0 | - |
| `vertiv_ac_system_monitor_baud_rate` | 24 | Monitor baud rate | 4.0 | - |
| `vertiv_ac_system_monitor_address` | 25 | Monitor the address | 200.0 | - |
| `vertiv_ac_system_unit_count` | 106 | Number of units | 1.0 | - |
| `vertiv_ac_system_main_delay_minutes` | 99 | Main delay | 2.0 | Min |
| `vertiv_ac_system_low_pressure_alarm_delay_seconds` | 121 | Low pressure alarm delay | 360.0 | Sec |
| `vertiv_ac_system_short_cycle_alarm_times_per_hour` | 122 | Short cycle alarm value | 4.0 | Times/Hour |
| `vertiv_ac_system_exhaust_superheat_low_alarm_delay_sec` | 124 | Exhaust superheat low alarm delay | 360.0 | Sec |

---

## 十二、解析规范

### 枚举值 → Gauge 数值映射

```
格式：EnumName[N]  →  取括号内整数 N 作为 Gauge 值

Running[0]        → 0
In[0]             → 0
NotIn[1]          → 1
TurnON[1]         → 1
Stop[0]           → 0
Enable[1]         → 1
Disable[0]        → 0
OutputMove[1]     → 1
OutputClose[0]    → 0
Open[1]           → 1
Close[0]          → 0
Normal[0]         → 0
Have[1]           → 1
GenerateNewAlarm[1] → 1
```

### 数值型字段解析规则

```
格式：field_id,字段名,数值,单位,timestamp,flag1,flag2,flag3,type,subtype

- 第3列：float64 数值直接映射为 Gauge
- 第4列：单位用于指标名后缀选择
- 第9列 type=2 → 数值型；type=5 → 枚举型
- 第9列 type=1 → 运行状态（枚举）
```

### 指标类型决策规则

| 条件 | Prometheus 类型 |
|------|----------------|
| 实时测量值（温度/湿度/压力/电压） | Gauge |
| 设定值/阈值 | Gauge |
| 累计运行小时数 | Counter（命名含 `_total`） |
| 累计次数（告警历史/启停记录） | Counter（命名含 `_total`） |
| 开关/状态（0或1） | Gauge |
| 告警属性（是否启用） | Gauge |
