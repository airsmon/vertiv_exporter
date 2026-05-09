# Vertiv THD 温湿度传感器 Prometheus 指标完整列表

> 数据来源：`ENV_THD / THD_SENSOR` 设备（equip_id: 5005）  
> 设备类型：机柜冷/热通道温湿度传感器 + 通道门状态  
> Labels（所有指标共有）：`instance="dc-rack-01"` `device="ENV_THD"` `equip_id="5005"`  
> 命名规范：`vertiv_thd_<category>_<name>_<unit>`

---

## 设计说明：Label 化处理

THD 数据的显著特点是**多机柜、多位置、同类字段大量重复**。  
若每个字段单独一条指标（如 `vertiv_thd_rack1_cool_aisle_top_temp_celsius`），会产生大量冗余 Desc。

**推荐方案：用 Label 区分机柜与位置，同类指标合并为一条。**

```
vertiv_thd_temperature_celsius{rack="RACK1", aisle="cool", position="top"}   21.7
vertiv_thd_temperature_celsius{rack="RACK1", aisle="cool", position="middle"} 22.9
vertiv_thd_temperature_celsius{rack="RACK1", aisle="hot",  position="top"}   29.4
```

额外 Labels：

| Label | 可选值 | 说明 |
|-------|--------|------|
| `rack` | `RACK1` `RACK2` `RACK3` `RACK4` `RACK5` `RACK_PMC` | 机柜标识 |
| `aisle` | `cool` `hot` | 冷通道 / 热通道 |
| `position` | `top` `middle` `bottom` | 传感器位置（上/中/下） |

---

## 一、通道温度 `vertiv_thd_temperature_celsius` (Gauge)

> 字段名规律：`RACK{N} {Cool|Hot} Aisle {Top|Middle|Bottom} Temp`

| 指标（含 Labels） | 字段 ID | 示例值(℃) |
|------------------|---------|-----------|
| `{rack="RACK1", aisle="cool", position="top"}` | 3 | 21.7 |
| `{rack="RACK1", aisle="cool", position="middle"}` | 4 | 22.9 |
| `{rack="RACK1", aisle="cool", position="bottom"}` | 6 | 19.4 |
| `{rack="RACK1", aisle="hot", position="top"}` | 7 | 29.4 |
| `{rack="RACK1", aisle="hot", position="middle"}` | 8 | 31.4 |
| `{rack="RACK1", aisle="hot", position="bottom"}` | 10 | 26.4 |
| `{rack="RACK2", aisle="cool", position="top"}` | 22 | 24.5 |
| `{rack="RACK2", aisle="cool", position="middle"}` | 23 | 23.5 |
| `{rack="RACK2", aisle="cool", position="bottom"}` | 25 | 20.9 |
| `{rack="RACK2", aisle="hot", position="top"}` | 26 | 28.9 |
| `{rack="RACK2", aisle="hot", position="middle"}` | 27 | 31.4 |
| `{rack="RACK2", aisle="hot", position="bottom"}` | 29 | 29.4 |
| `{rack="RACK3", aisle="cool", position="top"}` | 41 | 26.2 |
| `{rack="RACK3", aisle="cool", position="middle"}` | 42 | 23.4 |
| `{rack="RACK3", aisle="cool", position="bottom"}` | 44 | 24.0 |
| `{rack="RACK3", aisle="hot", position="top"}` | 45 | 28.3 |
| `{rack="RACK3", aisle="hot", position="middle"}` | 46 | 27.2 |
| `{rack="RACK3", aisle="hot", position="bottom"}` | 48 | 26.4 |
| `{rack="RACK4", aisle="cool", position="top"}` | 60 | 23.6 |
| `{rack="RACK4", aisle="cool", position="middle"}` | 61 | 21.8 |
| `{rack="RACK4", aisle="cool", position="bottom"}` | 63 | 22.1 |
| `{rack="RACK4", aisle="hot", position="top"}` | 64 | 29.5 |
| `{rack="RACK4", aisle="hot", position="middle"}` | 65 | 27.0 |
| `{rack="RACK4", aisle="hot", position="bottom"}` | 67 | 27.8 |
| `{rack="RACK5", aisle="cool", position="top"}` | 79 | 20.7 |
| `{rack="RACK5", aisle="cool", position="middle"}` | 80 | 20.8 |
| `{rack="RACK5", aisle="cool", position="bottom"}` | 82 | 19.6 |
| `{rack="RACK5", aisle="hot", position="top"}` | 83 | 26.1 |
| `{rack="RACK5", aisle="hot", position="middle"}` | 84 | 24.6 |
| `{rack="RACK5", aisle="hot", position="bottom"}` | 86 | 26.2 |
| `{rack="RACK_PMC", aisle="cool", position="top"}` | 155 | 21.7 |
| `{rack="RACK_PMC", aisle="cool", position="middle"}` | 156 | 22.4 |
| `{rack="RACK_PMC", aisle="cool", position="bottom"}` | 158 | 19.0 |
| `{rack="RACK_PMC", aisle="hot", position="top"}` | 159 | 26.2 |
| `{rack="RACK_PMC", aisle="hot", position="middle"}` | 160 | 25.2 |
| `{rack="RACK_PMC", aisle="hot", position="bottom"}` | 162 | 24.8 |

**Prometheus 文本格式示例：**
```
# HELP vertiv_thd_temperature_celsius Aisle temperature measurement
# TYPE vertiv_thd_temperature_celsius gauge
vertiv_thd_temperature_celsius{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="cool",position="top"} 21.7
vertiv_thd_temperature_celsius{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="cool",position="middle"} 22.9
vertiv_thd_temperature_celsius{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="hot",position="top"} 29.4
```

---

## 二、通道湿度 `vertiv_thd_humidity_percent` (Gauge)

> 每个机柜冷/热通道各一个湿度传感器（无 position 区分）

| 指标（含 Labels） | 字段 ID | 示例值(%) |
|------------------|---------|----------|
| `{rack="RACK1", aisle="cool"}` | 5 | 40.4 |
| `{rack="RACK1", aisle="hot"}` | 9 | 24.4 |
| `{rack="RACK2", aisle="cool"}` | 24 | 42.0 |
| `{rack="RACK2", aisle="hot"}` | 28 | 27.3 |
| `{rack="RACK3", aisle="cool"}` | 43 | 41.5 |
| `{rack="RACK3", aisle="hot"}` | 47 | 32.7 |
| `{rack="RACK4", aisle="cool"}` | 62 | 46.3 |
| `{rack="RACK4", aisle="hot"}` | 66 | 33.1 |
| `{rack="RACK5", aisle="cool"}` | 81 | 50.1 |
| `{rack="RACK5", aisle="hot"}` | 85 | 38.1 |
| `{rack="RACK_PMC", aisle="cool"}` | 157 | 44.8 |
| `{rack="RACK_PMC", aisle="hot"}` | 161 | 37.6 |

**Prometheus 文本格式示例：**
```
# HELP vertiv_thd_humidity_percent Aisle humidity measurement
# TYPE vertiv_thd_humidity_percent gauge
vertiv_thd_humidity_percent{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="cool"} 40.4
vertiv_thd_humidity_percent{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="hot"} 24.4
```

---

## 三、通道门状态 `vertiv_thd_door_status` (Gauge)

> `Normal[0]` → 0（正常关闭），`Open[1]` → 1（门开/异常）  
> 可用于触发 Alertmanager 告警规则

| 指标（含 Labels） | 字段 ID | 示例值 |
|------------------|---------|--------|
| `{rack="RACK1", aisle="cool"}` | 12 | Normal[0] → 0 |
| `{rack="RACK1", aisle="hot"}` | 16 | Normal[0] → 0 |
| `{rack="RACK2", aisle="cool"}` | 31 | Normal[0] → 0 |
| `{rack="RACK2", aisle="hot"}` | 35 | Normal[0] → 0 |
| `{rack="RACK3", aisle="cool"}` | 50 | Normal[0] → 0 |
| `{rack="RACK3", aisle="hot"}` | 54 | Normal[0] → 0 |
| `{rack="RACK4", aisle="cool"}` | 69 | Normal[0] → 0 |
| `{rack="RACK4", aisle="hot"}` | 73 | Normal[0] → 0 |
| `{rack="RACK5", aisle="cool"}` | 88 | Normal[0] → 0 |
| `{rack="RACK5", aisle="hot"}` | 92 | Normal[0] → 0 |
| `{rack="RACK_PMC", aisle="cool"}` | 164 | Normal[0] → 0 |
| `{rack="RACK_PMC", aisle="hot"}` | 168 | Normal[0] → 0 |

**Prometheus 文本格式示例：**
```
# HELP vertiv_thd_door_status Aisle door status (0=Normal, 1=Open/Abnormal)
# TYPE vertiv_thd_door_status gauge
vertiv_thd_door_status{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="cool"} 0
vertiv_thd_door_status{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1",aisle="hot"} 0
```

---

## 四、THD 传感器通信状态 `vertiv_thd_comm_status` (Gauge)

> `Normal[0]` → 0（通信正常），`Fault[1]` → 1（通信故障）

| 指标（含 Labels） | 字段 ID | 示例值 |
|------------------|---------|--------|
| `{rack="RACK1"}` | 3010 | Normal[0] → 0 |
| `{rack="RACK2"}` | 3011 | Normal[0] → 0 |
| `{rack="RACK3"}` | 3012 | Normal[0] → 0 |
| `{rack="RACK4"}` | 3013 | Normal[0] → 0 |
| `{rack="RACK5"}` | 3014 | Normal[0] → 0 |
| `{rack="RACK_PMC"}` | 3018 | Normal[0] → 0 |

**Prometheus 文本格式示例：**
```
# HELP vertiv_thd_comm_status THD sensor communication status (0=Normal, 1=Fault)
# TYPE vertiv_thd_comm_status gauge
vertiv_thd_comm_status{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK1"} 0
vertiv_thd_comm_status{instance="dc-rack-01",device="ENV_THD",equip_id="5005",rack="RACK_PMC"} 0
```

---

## 五、机柜编号映射 `vertiv_thd_rack_id` (Gauge)

> 字段 10000~10008：记录各机柜的系统内部编号，可用于关联其他子系统数据

| 指标（含 Labels） | 字段 ID | 值 |
|------------------|---------|---|
| `{rack="RACK1"}` | 10000 | 1 |
| `{rack="RACK2"}` | 10001 | 2 |
| `{rack="RACK3"}` | 10002 | 3 |
| `{rack="RACK4"}` | 10003 | 4 |
| `{rack="RACK5"}` | 10004 | 5 |
| `{rack="RACK_PMC"}` | 10008 | 9 |

---

## 六、全局告警统计 `vertiv_thd_high_temp_alarm_rack_count` (Gauge)

> 当前触发高温告警的机柜数量，字段 ID 10012

```
# HELP vertiv_thd_high_temp_alarm_rack_count Number of racks with active high temperature alarm
# TYPE vertiv_thd_high_temp_alarm_rack_count gauge
vertiv_thd_high_temp_alarm_rack_count{instance="dc-rack-01",device="ENV_THD",equip_id="5005"} 0
```

---

## 七、指标汇总

| 指标名 | 类型 | Label 维度 | 时间序列数 |
|--------|------|-----------|-----------|
| `vertiv_thd_temperature_celsius` | Gauge | rack × aisle × position | 36 |
| `vertiv_thd_humidity_percent` | Gauge | rack × aisle | 12 |
| `vertiv_thd_door_status` | Gauge | rack × aisle | 12 |
| `vertiv_thd_comm_status` | Gauge | rack | 6 |
| `vertiv_thd_rack_id` | Gauge | rack | 6 |
| `vertiv_thd_high_temp_alarm_rack_count` | Gauge | 无 | 1 |
| **合计** | | | **73 条时间序列** |

---

## 八、Alertmanager 告警规则建议

```yaml
groups:
  - name: vertiv_thd_alerts
    rules:
      # 热通道顶部温度过高
      - alert: HotAisleTopTempHigh
        expr: vertiv_thd_temperature_celsius{aisle="hot", position="top"} > 35
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.rack }} 热通道顶部温度过高: {{ $value }}℃"

      # 冷通道温度过高（空调制冷效果下降）
      - alert: CoolAisleTempHigh
        expr: vertiv_thd_temperature_celsius{aisle="cool"} > 25
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.rack }} 冷通道温度异常: {{ $value }}℃"

      # 机柜门被打开
      - alert: AisleDoorOpen
        expr: vertiv_thd_door_status == 1
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.rack }} {{ $labels.aisle }} 通道门异常打开"

      # THD 传感器通信故障
      - alert: THDCommFault
        expr: vertiv_thd_comm_status == 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.rack }} THD 传感器通信故障"
```
