# Vertiv UPS Prometheus 指标完整列表

> 数据来源：`ENP_UPS_ITA2[COM]` 设备（equip_id: 491，device: UPS_1）  
> Labels（所有指标共有）：`instance="dc-rack-01"` `device="UPS_1"` `equip_id="491"`  
> 命名规范：`vertiv_ups_<category>_<name>_<unit>`  
> **多相电气指标** 用 `phase` Label 合并（`phase="A"` `phase="B"` `phase="C"`）

---

## 一、输入电压 `vertiv_ups_input_voltage_volts` (Gauge)

> 相电压与线电压分两个指标，用 `phase` / `line` label 区分

### 1.1 相电压 — `vertiv_ups_input_phase_voltage_volts`

| Labels | 字段 ID | 示例值(V) |
|--------|---------|----------|
| `{phase="A"}` | 2 | 217.8 |
| `{phase="B"}` | 3 | 219.6 |
| `{phase="C"}` | 4 | 221.9 |

### 1.2 线电压 — `vertiv_ups_input_line_voltage_volts`

| Labels | 字段 ID | 示例值(V) |
|--------|---------|----------|
| `{line="AB"}` | 16 | 378.4 |
| `{line="BC"}` | 69 | 381.3 |
| `{line="CA"}` | 63 | 381.8 |

### 1.3 输入电流 — `vertiv_ups_input_current_amperes`

| Labels | 字段 ID | 示例值(A) |
|--------|---------|----------|
| `{phase="A"}` | 124 | 10.3 |
| `{phase="B"}` | 125 | 9.9 |
| `{phase="C"}` | 126 | 10.4 |

### 1.4 输入频率 — `vertiv_ups_input_frequency_hz` (Gauge, 无 Labels)

| 字段 ID | 示例值(HZ) |
|---------|-----------|
| 12 | 50.0 |

### 1.5 输入功率因数 — `vertiv_ups_input_power_factor`

| Labels | 字段 ID | 示例值 |
|--------|---------|--------|
| `{phase="A"}` | 23 | 0.99 |
| `{phase="B"}` | 71 | 0.99 |
| `{phase="C"}` | 77 | 0.99 |

---

## 二、旁路电压 `vertiv_ups_bypass_*` (Gauge)

### 2.1 旁路相电压 — `vertiv_ups_bypass_phase_voltage_volts`

| Labels | 字段 ID | 示例值(V) |
|--------|---------|----------|
| `{phase="A"}` | 143 | 221.5 |
| `{phase="B"}` | 144 | 220.0 |
| `{phase="C"}` | 145 | 223.4 |

### 2.2 旁路线电压 — `vertiv_ups_bypass_line_voltage_volts`

| Labels | 字段 ID | 示例值(V) |
|--------|---------|----------|
| `{line="AB"}` | 29 | 382.2 |
| `{line="BC"}` | 30 | 384.3 |
| `{line="CA"}` | 31 | 386.2 |

### 2.3 旁路频率 — `vertiv_ups_bypass_frequency_hz` (Gauge, 无 Labels)

| 字段 ID | 示例值(HZ) |
|---------|-----------|
| 146 | 49.97 |

---

## 三、输出电气类 (Gauge)

### 3.1 输出相电压 — `vertiv_ups_output_phase_voltage_volts`

| Labels | 字段 ID | 示例值(V) |
|--------|---------|----------|
| `{phase="A"}` | 5 | 219.7 |
| `{phase="B"}` | 6 | 221.2 |
| `{phase="C"}` | 7 | 218.7 |

### 3.2 输出电流 — `vertiv_ups_output_current_amperes`

| Labels | 字段 ID | 示例值(A) |
|--------|---------|----------|
| `{phase="A"}` | 8 | 15.2 |
| `{phase="B"}` | 9 | 2.8 |
| `{phase="C"}` | 10 | 15.0 |

### 3.3 输出频率 — `vertiv_ups_output_frequency_hz` (Gauge, 无 Labels)

| 字段 ID | 示例值(HZ) |
|---------|-----------|
| 11 | 49.97 |

### 3.4 输出功率因数 — `vertiv_ups_output_power_factor`

| Labels | 字段 ID | 示例值 |
|--------|---------|--------|
| `{phase="A"}` | 35 | 0.98 |
| `{phase="B"}` | 36 | 0.94 |
| `{phase="C"}` | 37 | 0.99 |

---

## 四、输出功率类 (Gauge)

> 本机（Local）与并机系统（System）分两组，`scope` label 区分

### 4.1 输出有功功率 — `vertiv_ups_output_active_power_kilowatts`

| Labels | 字段 ID | 示例值(KW) |
|--------|---------|-----------|
| `{scope="local", phase="A"}` | 134 | 3.11 |
| `{scope="local", phase="B"}` | 135 | 0.44 |
| `{scope="local", phase="C"}` | 136 | 3.04 |
| `{scope="system", phase="A"}` | 54 | 3.13 |
| `{scope="system", phase="B"}` | 55 | 0.43 |
| `{scope="system", phase="C"}` | 56 | 3.03 |

### 4.2 输出视在功率 — `vertiv_ups_output_apparent_power_kva`

| Labels | 字段 ID | 示例值(KVA) |
|--------|---------|------------|
| `{scope="local", phase="A"}` | 44 | 3.15 |
| `{scope="local", phase="B"}` | 45 | 0.49 |
| `{scope="local", phase="C"}` | 46 | 3.06 |
| `{scope="system", phase="A"}` | 57 | 3.17 |
| `{scope="system", phase="B"}` | 58 | 0.49 |
| `{scope="system", phase="C"}` | 59 | 3.05 |

### 4.3 输出负载率 — `vertiv_ups_output_load_percent`

| Labels | 字段 ID | 示例值(%) |
|--------|---------|----------|
| `{phase="A"}` | 13 | 47.2 |
| `{phase="B"}` | 14 | 7.4 |
| `{phase="C"}` | 15 | 45.8 |

---

## 五、电池类 (Gauge)

| 指标名 | 字段 ID | 示例值 | 单位 | 说明 |
|--------|---------|--------|------|------|
| `vertiv_ups_battery_voltage_volts` | 18 | 219.1 | V | 正组电池电压 |
| `vertiv_ups_battery_negative_voltage_volts` | 66 | 218.6 | V | 负组电池电压 |
| `vertiv_ups_battery_charge_current_amperes` | 64 | 0.2 | A | 正组充电电流 |
| `vertiv_ups_battery_discharge_current_amperes` | 65 | 0.0 | A | 正组放电电流 |
| `vertiv_ups_battery_negative_charge_current_amperes` | 67 | 0.07 | A | 负组充电电流 |
| `vertiv_ups_battery_negative_discharge_current_amperes` | 68 | 0.0 | A | 负组放电电流 |
| `vertiv_ups_battery_capacity_percent` | 72 | 100.0 | % | 当前电量 SOC |
| `vertiv_ups_battery_backup_time_minutes` | 17 | 33.9 | Min | 估算后备时间 |
| `vertiv_ups_battery_discharging_time_seconds` | 167 | 3054.0 | s | 累计放电时长 |
| `vertiv_ups_battery_discharge_count_total` | 73 | 59.0 | - | 累计放电次数（Counter） |

---

## 六、累计电能类 (Counter)

> 长期单调递增，使用 **Counter** 类型，命名含 `_total`

| 指标名 | 字段 ID | 示例值 | 单位 |
|--------|---------|--------|------|
| `vertiv_ups_input_energy_kwh_total` | 75 | 261917.0 | KWH |
| `vertiv_ups_output_energy_kwh_total` | 76 | 251038.0 | KWH |

---

## 七、环境与系统类 (Gauge)

| 指标名 | 字段 ID | 示例值 | 单位 |
|--------|---------|--------|------|
| `vertiv_ups_ambient_temperature_celsius` | 24 | 20.5 | ℃ |
| `vertiv_ups_running_time_days` | 62 | 2413.0 | day |
| `vertiv_ups_parallel_machine_count` | 60 | 1.0 | - |

---

## 八、运行状态类 `vertiv_ups_status_*` (Gauge, 枚举取括号内整数)

| 指标名 | 字段 ID | 原始枚举值 | Gauge值 | 说明 |
|--------|---------|-----------|---------|------|
| `vertiv_ups_status_power_supply` | 25 | Utility Online[1] | **1** | 0=Battery/Bypass, 1=Utility Online |
| `vertiv_ups_status_input_power` | 79 | Utility Online[0] | **0** | 0=Online, 1=异常 |
| `vertiv_ups_status_battery` | 27 | Float Charging[1] | **1** | 浮充=1，放电=其他值 |
| `vertiv_ups_status_battery_negative_group` | 81 | Float Charging[1] | **1** | 负组电池状态 |
| `vertiv_ups_status_charger` | 82 | Charger On[0] | **0** | 0=充电器开启 |
| `vertiv_ups_status_parallel_system_power` | 83 | Main Inverter Power Supply[0] | **0** | 并机供电状态 |
| `vertiv_ups_status_inner_network` | 84 | Disconnected[1] | **1** | 1=内网断开 |
| `vertiv_ups_status_communication` | 456 | Normal[0] | **0** | 0=通信正常 |
| `vertiv_ups_status_input_phase_number` | 49 | Three Phase[3] | **3** | 输入相数 |
| `vertiv_ups_status_output_phase_number` | 34 | Three Phase[3] | **3** | 输出相数 |

---

## 九、Prometheus 文本格式示例（完整片段）

```
# HELP vertiv_ups_output_load_percent UPS output load percentage per phase
# TYPE vertiv_ups_output_load_percent gauge
vertiv_ups_output_load_percent{instance="dc-rack-01",device="UPS_1",equip_id="491",phase="A"} 47.2
vertiv_ups_output_load_percent{instance="dc-rack-01",device="UPS_1",equip_id="491",phase="B"} 7.4
vertiv_ups_output_load_percent{instance="dc-rack-01",device="UPS_1",equip_id="491",phase="C"} 45.8

# HELP vertiv_ups_battery_capacity_percent UPS battery state of charge
# TYPE vertiv_ups_battery_capacity_percent gauge
vertiv_ups_battery_capacity_percent{instance="dc-rack-01",device="UPS_1",equip_id="491"} 100.0

# HELP vertiv_ups_battery_backup_time_minutes Estimated UPS battery backup time
# TYPE vertiv_ups_battery_backup_time_minutes gauge
vertiv_ups_battery_backup_time_minutes{instance="dc-rack-01",device="UPS_1",equip_id="491"} 33.9

# HELP vertiv_ups_output_energy_kwh_total Total UPS output energy
# TYPE vertiv_ups_output_energy_kwh_total counter
vertiv_ups_output_energy_kwh_total{instance="dc-rack-01",device="UPS_1",equip_id="491"} 251038.0

# HELP vertiv_ups_status_power_supply UPS power supply status (1=Utility Online)
# TYPE vertiv_ups_status_power_supply gauge
vertiv_ups_status_power_supply{instance="dc-rack-01",device="UPS_1",equip_id="491"} 1
```

---

## 十、指标汇总

| 类别 | 指标名数量 | 时间序列数 | 类型 |
|------|-----------|-----------|------|
| 输入电压（相/线） | 2 | 6 | Gauge |
| 输入电流 | 1 | 3 | Gauge |
| 输入频率 | 1 | 1 | Gauge |
| 输入功率因数 | 1 | 3 | Gauge |
| 旁路电压（相/线） | 2 | 6 | Gauge |
| 旁路频率 | 1 | 1 | Gauge |
| 输出电压 | 1 | 3 | Gauge |
| 输出电流 | 1 | 3 | Gauge |
| 输出频率 | 1 | 1 | Gauge |
| 输出功率因数 | 1 | 3 | Gauge |
| 输出有功功率（本机/系统） | 1 | 6 | Gauge |
| 输出视在功率（本机/系统） | 1 | 6 | Gauge |
| 输出负载率 | 1 | 3 | Gauge |
| 电池（电压/电流/SOC/后备时间等） | 9 | 9 | Gauge |
| 电池放电次数 | 1 | 1 | Counter |
| 累计电能（输入/输出） | 2 | 2 | Counter |
| 环境与系统 | 3 | 3 | Gauge |
| 运行状态枚举 | 10 | 10 | Gauge |
| **合计** | **40** | **70** | — |

---

## 十一、Alertmanager 告警规则建议

```yaml
groups:
  - name: vertiv_ups_alerts
    rules:
      # UPS 切换到电池供电
      - alert: UPSOnBattery
        expr: vertiv_ups_status_power_supply != 1
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.device }} 已切换到电池/旁路供电"

      # 电池电量不足
      - alert: UPSBatteryLow
        expr: vertiv_ups_battery_capacity_percent < 30
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.device }} 电池电量不足: {{ $value }}%"

      # 后备时间过短
      - alert: UPSBackupTimeLow
        expr: vertiv_ups_battery_backup_time_minutes < 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.device }} 后备时间仅剩 {{ $value }} 分钟"

      # 单相负载过高（负载不均衡）
      - alert: UPSPhaseOverload
        expr: vertiv_ups_output_load_percent > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.device }} {{ $labels.phase }} 相负载过高: {{ $value }}%"

      # UPS 通信故障
      - alert: UPSCommFault
        expr: vertiv_ups_status_communication != 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.device }} 通信状态异常"
```
