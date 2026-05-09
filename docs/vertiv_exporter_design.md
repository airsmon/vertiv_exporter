# Vertiv 模块机柜 Prometheus Exporter 设计方案

## 1. 项目概述

本项目旨在为 Vertiv 模块化机柜（含精密空调 AC 等设备）开发一个符合 Prometheus 官方规范的 Exporter，使用 Golang 实现，支持多设备并发采集与 Prometheus 指标暴露。

- **采集目标**：Vertiv Web 管理界面（CGI 接口）
- **暴露方式**：HTTP `/metrics` 端点（标准 Prometheus 文本格式）
- **开发语言**：Go 1.21+
- **核心依赖**：`github.com/prometheus/client_golang`

---

## 2. 目录结构

```
vertiv_exporter/
├── cmd/
│   └── vertiv_exporter/
│       └── main.go              # 程序入口，flag 解析，HTTP 服务启动
├── internal/
│   ├── client/
│   │   ├── client.go            # HTTP 客户端，Session 管理，登录，心跳
│   │   └── parser.go            # CGI 响应解析（Y|... 格式）
│   ├── collector/
│   │   ├── collector.go         # 实现 prometheus.Collector 接口
│   │   └── metrics.go           # 所有 Desc（指标描述符）定义
│   └── config/
│       └── config.go            # YAML 配置读取与结构体定义
├── config.example.yaml          # 配置文件示例
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## 3. 配置文件设计（config.yaml）

```yaml
exporter:
  listen_address: ":9101"     # Exporter 监听地址
  metrics_path: "/metrics"    # 指标暴露路径
  scrape_timeout: 10s         # 单次采集超时

targets:
  - name: "dc-rack-01"        # 实例标签名，对应 instance label
    host: "https://192.168.1.100"
    username: "admin_encoded" # URL 编码后的账号
    password: "pwd_encoded"   # URL 编码后的密码
    tls_skip_verify: true     # 是否跳过证书校验
    devices:
      - name: "AC1"
        equip_id: 23
        elements: "16|32;700;0,2@33;700;0,157@..."
      - name: "AC2"
        equip_id: 24
        elements: "16|48;700;0,2@49;700;0,157@..."

  - name: "dc-rack-02"
    host: "https://192.168.1.101"
    username: "admin_encoded"
    password: "pwd_encoded"
    tls_skip_verify: true
    devices:
      - name: "AC1"
        equip_id: 23
        elements: "16|32;700;0,2@..."
```

---

## 4. 核心接口设计

### 4.1 HTTP 客户端（internal/client/client.go）

```go
type Client struct {
    host       string
    httpClient *http.Client
    jar        *cookiejar.Jar
    mu         sync.Mutex
}

// 构造函数：初始化 TLS 配置、CookieJar、预置语言 Cookie
func NewClient(host string, skipTLS bool) (*Client, error)

// 登录：POST /cgi-bin/login.cgi，写入 Session Cookie
func (c *Client) Login(username, password string) error

// 心跳保活：GET /cgi-bin/main_page_polling.cgi
func (c *Client) KeepAlive(ctx context.Context) error

// 数据采集：GET /cgi-bin/p101_refresh_page.cgi
func (c *Client) FetchDeviceData(ctx context.Context, equipID int, elements string) (map[string]float64, error)
```

### 4.2 响应解析（internal/client/parser.go）

解析 Vertiv CGI 的特殊格式响应：

```
Y|32~700~进/回风温差;700;0,2@5.3|33~700~回风温度;700;0,157@24.6|...
```

解析规则：
1. 去除前缀 `Y|`
2. 以 `|` 分割每条记录
3. 每条记录以 `;` 分割为 `meta;value`
4. `meta` 以 `~` 分割，取第 3 段为 metric 名称
5. 过滤含 `.gif` 的记录
6. 数值转换为 `float64`

```go
func ParseResponse(raw string) (map[string]float64, error)
```

### 4.3 Prometheus Collector（internal/collector/collector.go）

```go
type VertivCollector struct {
    config    *config.Config
    clients   map[string]*client.Client   // key: target.name
    descs     map[string]*prometheus.Desc
    scrapeDur prometheus.Histogram
    scrapeErr prometheus.Counter
}

// 实现 prometheus.Collector 接口
func (c *VertivCollector) Describe(ch chan<- *prometheus.Desc)
func (c *VertivCollector) Collect(ch chan<- prometheus.Metric)
```

`Collect` 并发采集所有 target，汇总后写入 channel。

---

## 5. 指标设计

所有指标遵循 Prometheus 命名规范：`vertiv_<subsystem>_<metric_name>[_unit]`

### 5.1 空调设备指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `vertiv_ac_return_air_temp_celsius` | Gauge | 回风温度 (°C) |
| `vertiv_ac_supply_air_temp_celsius` | Gauge | 送风温度 (°C) |
| `vertiv_ac_air_temp_diff_celsius` | Gauge | 进/回风温差 (°C) |
| `vertiv_ac_humidity_percent` | Gauge | 环境湿度 (%) |
| `vertiv_ac_compressor_status` | Gauge | 压缩机状态 (0=关, 1=开) |
| `vertiv_ac_fan_speed_rpm` | Gauge | 风机转速 (RPM) |
| `vertiv_ac_cooling_capacity_kw` | Gauge | 制冷量 (kW) |
| `vertiv_ac_power_consumption_kw` | Gauge | 功耗 (kW) |
| `vertiv_ac_alarm_status` | Gauge | 告警状态 (0=正常, 1=告警) |

### 5.2 Exporter 自身指标（标准规范要求）

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `vertiv_exporter_scrape_duration_seconds` | Histogram | 每次采集耗时 |
| `vertiv_exporter_scrape_errors_total` | Counter | 累计采集失败次数 |
| `vertiv_exporter_up` | Gauge | 目标在线状态 (0/1) |

### 5.3 Label 设计

所有设备指标携带以下 Labels：

```
instance="dc-rack-01"    # 对应 config.targets[].name
device="AC1"             # 对应 config.targets[].devices[].name
equip_id="23"            # 设备 ID
```

---

## 6. Session 管理与保活机制

Vertiv 设备的 CGI 接口需要维持登录 Session，设计以下保活策略：

```
┌──────────────────────────────────────────────┐
│  每个 Target 独立的 goroutine                 │
│                                              │
│  1. 启动时登录，获取 Session Cookie           │
│  2. 每 30s 发送心跳 keep_alive               │
│  3. 采集时检测 401/302，自动重新登录          │
│  4. 使用 sync.RWMutex 保护 Cookie 并发访问   │
└──────────────────────────────────────────────┘
```

心跳 goroutine 伪代码：

```go
func (c *Client) StartKeepAlive(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := c.KeepAlive(ctx); err != nil {
                // 心跳失败时触发重新登录
                _ = c.Login(c.username, c.password)
            }
        }
    }
}
```

---

## 7. 并发采集设计

`Collect()` 采用 `errgroup` + `sync.WaitGroup` 并发拉取所有 target 的所有设备：

```go
func (col *VertivCollector) Collect(ch chan<- prometheus.Metric) {
    var wg sync.WaitGroup
    for _, target := range col.config.Targets {
        wg.Add(1)
        go func(t config.Target) {
            defer wg.Done()
            for _, dev := range t.Devices {
                data, err := col.clients[t.Name].FetchDeviceData(ctx, dev.EquipID, dev.Elements)
                if err != nil {
                    // 上报 scrape_errors_total++，vertiv_up=0
                    return
                }
                col.emitMetrics(ch, t.Name, dev, data)
            }
        }(target)
    }
    wg.Wait()
}
```

---

## 8. main.go 入口

```go
func main() {
    configFile := flag.String("config.file", "config.yaml", "配置文件路径")
    flag.Parse()

    cfg := config.Load(*configFile)
    collector := collector.New(cfg)
    prometheus.MustRegister(collector)

    http.Handle(cfg.Exporter.MetricsPath, promhttp.Handler())
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, `<html><body>
            <h1>Vertiv Exporter</h1>
            <a href="%s">Metrics</a></body></html>`,
            cfg.Exporter.MetricsPath)
    })

    log.Printf("Listening on %s", cfg.Exporter.ListenAddress)
    log.Fatal(http.ListenAndServe(cfg.Exporter.ListenAddress, nil))
}
```

---

## 9. Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o vertiv_exporter ./cmd/vertiv_exporter

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/vertiv_exporter .
COPY config.example.yaml ./config.yaml
EXPOSE 9101
ENTRYPOINT ["./vertiv_exporter", "--config.file=/app/config.yaml"]
```

---

## 10. Prometheus 配置示例

```yaml
scrape_configs:
  - job_name: "vertiv"
    static_configs:
      - targets: ["localhost:9101"]
    scrape_interval: 30s
    scrape_timeout: 15s
```

---

## 11. 符合 Prometheus Exporter 规范的关键点

| 规范要求 | 本方案实现方式 |
|----------|----------------|
| 指标命名含 namespace | `vertiv_` 前缀统一 |
| 使用 Gauge/Counter/Histogram | 状态用 Gauge，错误计数用 Counter，耗时用 Histogram |
| `/metrics` 端点 | `promhttp.Handler()` 标准实现 |
| 暴露 `_up` 指标 | `vertiv_exporter_up` per target |
| 暴露 `scrape_duration_seconds` | `vertiv_exporter_scrape_duration_seconds` Histogram |
| 并发安全 | `sync.RWMutex` 保护 Session |
| 优雅退出 | `context.WithCancel` 传播取消信号 |
| 支持多 target | config.yaml 多 target 配置 |



浏览器抓取数据：

AC_1
https://vertiv.cn-sh-1.daocloud.io/cgi-bin/p05_equip_sample.cgi?sand=0.9452394395827178&_equipId=23&_op_type=1

3021,AC_1,ENP_AC_SRVII[COM]^2,Return air temperature measurement,28.600000,℃,1778314683,0,1,1,2,2;3,Return air humidity measurement,30.400000,%,1778314683,0,1,1,2,2;9,Air temperature measurement,30.200001,℃,1778314683,0,1,1,2,2;10,Exhaust temperature measurement,55.000000,℃,1778314683,0,1,1,2,2;324,Inspiratory temperature measurement,21.200001,℃,1778314683,0,1,1,2,2;157,Air supply temperature 1 measured value,19.400000,℃,1778314683,0,1,1,2,2;263,Air supply temperature 2 measured value,18.299999,℃,1778314683,0,1,1,2,2;323,Supply Air temperature setting,20.000000,℃,1778314683,0,1,1,2,2;325,Mean temperature measurement of Supply Air,18.799999,℃,1778314683,0,1,1,2,2;15,Air supply temperature 1 correction value,0.000000,℃,1778314683,0,1,1,2,2;16,Air supply temperature 2 correction value,0.000000,℃,1778314683,0,1,1,2,2;17,Return air temperature correction value,0.000000,℃,1778314683,0,1,1,2,2;18,Airflow temperature correction value,0.000000,℃,1778314683,0,1,1,2,2;19,Exhaust temperature correction value,0.000000,℃,1778314683,0,1,1,2,2;20,Inspiratory temperature correction value,0.000000,℃,1778314683,0,1,1,2,2;4,Humidity setting,50.000000,%,1778314683,0,1,1,2,2;65,Humidity ratio,5.000000,%,1778314683,0,1,1,2,2;21,Return air humidity correction value,0.000000,%,1778314683,0,1,1,2,2;326,Exhaust pressure measurement,20.100000,Bar,1778314683,0,1,1,2,2;14,Inspiratory pressure measurement,10.600000,Bar,1778314683,0,1,1,2,2;22,Exhaust pressure correction value,0.000000,Bar,1778314683,0,1,1,2,2;23,Inspiratory pressure correction value,0.000000,Bar,1778314683,0,1,1,2,2;26,Phase A voltage,213.800003,V,1778314683,0,1,1,2,2;27,B phase voltage,214.899994,V,1778314683,0,1,1,2,2;28,C phase voltage,214.899994,V,1778314683,0,1,1,2,2;29,Power frequency,49.900002,HZ,1778314683,0,1,1,2,2;30,Compressor shortest running time,15.000000,Min,1778314683,0,1,1,2,2;31,Compressor shortest downtime,2.000000,Min,1778314683,0,1,1,2,2;34,Inspiratory evaporation temperature,12.200000,℃,1778314683,0,1,1,2,2;35,Exhaust condensing temperature,34.400002,℃,1778314683,0,1,1,2,2;36,Inspiratory superheat,9.000000,℃,1778314683,0,1,1,2,2;37,Exhaust superheat,20.600000,℃,1778314683,0,1,1,2,2;38,Theoretical air supply humidity,51.700001,%,1778314683,0,1,1,2,2;39,Current air supply humidity,55.700001,%,1778314683,0,1,1,2,2;32,Fan start delay,10.000000,Sec,1778314683,0,1,1,2,2;33,Fan downtime,30.000000,Sec,1778314683,0,1,1,2,2;24,Monitor baud rate,4.000000,,1778314683,0,1,1,2,2;25,Monitor the address,200.000000,,1778314683,0,1,1,2,2;40,DIP switch value,0.000000,,1778314683,0,1,1,2,2;41,Number of alarm states,1.000000,,1778314683,0,1,1,2,2;42,Number of alarm history,153.000000,,1778314683,0,1,1,2,2;43,Number of compressor start and stop records,50.000000,,1778314683,0,1,1,2,2;44,Number of pump start and stop records,50.000000,,1778314683,0,1,1,2,2;45,Number of fan start and stop records,50.000000,,1778314683,0,1,1,2,2;46,Number of electric heating start and stop records,1.000000,,1778314683,0,1,1,2,2;47,Number of humidifier start and stop records,0.000000,,1778314683,0,1,1,2,2;48,Compressor running hours,0.000000,Hour,1778314683,0,0,1,2,2;49,Pump running hours,1026.000000,Hour,1778314683,0,1,1,2,2;50,Fan running hours,0.000000,Hour,1778314683,0,0,1,2,2;51,Electric heating operation hours,0.000000,Hour,1778314683,0,1,1,2,2;52,Humidifier running hours,0.000000,Hour,1778314683,0,1,1,2,2;53,Control board coding,0.000000,,1778314683,0,0,1,2,2;54,Control board serial number,0.000000,,1778314683,0,1,1,2,2;55,Software version is high,154.000000,,1778314683,0,1,1,2,2;56,The software version is low,0.000000,,1778314683,0,0,1,2,2;57,Monitoring protocol,0.000000,,1778314683,0,1,1,2,2;58,Start processing signs,1.000000,,1778314683,0,1,1,2,2;59,System time (years),2026.000000,,1778314683,0,1,1,2,2;60,System time (months),5.000000,,1778314683,0,1,1,2,2;61,System time (day),9.000000,,1778314683,0,1,1,2,2;62,System time (hours),8.000000,,1778314683,0,1,1,2,2;63,System time (minutes),17.000000,,1778314683,0,1,1,2,2;64,System time (seconds),53.000000,,1778314683,0,1,1,2,2;5,Air conditioning operation status,Running[0],,1778314683,0,1,1,1,5;66,High voltage alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;67,Low voltage alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;68,High Voltage Abnormal Alarm Attribute,TurnON[1],,1778314683,0,1,1,0,5;69,Exhaust high temperature alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;70,Exhaust superheat low alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;71,Return air temperature alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;72,Air temperature alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;73,Air temperature alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;74,Return air humidity alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;75,Return air low humidity alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;76,High voltage lock alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;77,Low-voltage lock alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;78,Exhaust high temperature lock alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;79,Exhaust superheat low lock alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;80,Power loss alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;81,Power overvoltage alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;82,Power undervoltage alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;83,Power Missing Alarm Attribute,TurnON[2],,1778314683,0,1,1,0,5;84,Floor overflow alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;85,High water alarm attribute,Stop[0],,1778314683,0,1,1,0,5;86,Filter plugging alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;87,Filter maintenance alert attribute,TurnON[2],,1778314683,0,1,1,0,5;88,Airflow loss alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;89,Low voltage sensor lock alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;90,Remote shutdown alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;3091,Group Control Host Loses Alarm Attribute,TurnON[1],,1778314683,0,1,1,0,5;92,Group Control Slave Loses Alarm Attributes,TurnON[1],,1778314683,0,1,1,0,5;93,Return air temperature sensor fault alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;94,Return air humidity sensor fault alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;95,Air temperature difference sensor fault alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;96,Air supply temperature sensor fault alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;97,Remote temperature sensor fault alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;98,High pressure sensor failure alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;99,Main delay,2.000000,Min,1778314683,0,1,1,2,2;100,First run unit,0.000000,,1778314683,0,1,1,2,2;101,Heating function,Disable[0],,1778314683,0,1,1,0,5;102,Number of remote temperature sensors,0.000000,,1778314683,0,1,1,2,2;103,Condensate pump,Have[1],,1778314683,0,1,1,0,5;104,Group control mode,Single[0],,1778314683,0,1,1,0,5;105,Unit address,0.000000,,1778314683,0,1,1,2,2;106,Number of units,1.000000,,1778314683,0,1,1,2,2;107,Number of machines,0.000000,,1778314683,0,1,1,2,2;108,Number of rounds,0.000000,,1778314683,0,1,1,2,2;109,Round patrol moment,12.000000,,1778314683,0,1,1,2,2;110,Round robin cycle,0.000000,,1778314683,0,1,1,2,2;111,Temperature dead zone,0.500000,℃,1778314683,0,1,1,2,2;112,Humidity dead zone,3.000000,%,1778314683,0,1,1,2,2;113,Remote temperature setting,20.000000,℃,1778314683,0,1,1,2,2;114,Return air temperature setting,30.000000,℃,1778314683,0,1,1,2,2;115,Compressor capacity actual value,31.000000,%,1778314683,0,1,1,2,2;116,Compressor capacity output value,31.000000,%,1778314683,0,1,1,2,2;117,Fan speed,40.000000,%,1778314683,0,1,1,2,2;118,Expansion valve opening degree,16.000000,%,1778314683,0,1,1,2,2;121,Low pressure alarm delay,360.000000,Sec,1778314683,0,1,1,2,2;122,Short cycle alarm value,4.000000,Times/Hour,1778314683,0,1,1,2,2;123,Filter maintenance reminder time,90.000000,Day,1778314683,0,1,1,2,2;124,Exhaust superheat low alarm delay,360.000000,Sec,1778314683,0,1,1,2,2;125,Floor overflow treatment,1.000000,,1778314683,0,1,1,2,2;126,Dehumidification run time,15.000000,Min,1778314683,0,1,1,2,2;127,Dehumidification stop temperature difference,-5.000000,℃,1778314683,0,1,1,2,2;128,Number of high pressure anomalies recorded,14.000000,,1778314683,0,1,1,2,2;130,Remote temperature 1 measurements,0.000000,℃,1778314683,0,0,1,2,2;131,Remote temperature 2 measurements,0.000000,℃,1778314683,0,0,1,2,2;132,Remote temperature 3 measurements,0.000000,℃,1778314683,0,0,1,2,2;133,Remote temperature 4 measurements,0.000000,℃,1778314683,0,0,1,2,2;134,Remote temperature 5 measurements,0.000000,℃,1778314683,0,0,1,2,2;135,Remote temperature 6 measurements,0.000000,℃,1778314683,0,0,1,2,2;136,Remote temperature 7 measurements,0.000000,℃,1778314683,0,0,1,2,2;137,Remote temperature 8 measurements,0.000000,℃,1778314683,0,0,1,2,2;138,Remote temperature 9 measurements,0.000000,℃,1778314683,0,0,1,2,2;139,Remote temperature10 measurements,0.000000,℃,1778314683,0,0,1,2,2;140,Remote average temperature,0.000000,℃,1778314683,0,0,1,2,2;141,Air temperature alarm value,30.000000,℃,1778314683,0,1,1,2,2;142,Low temperature alarm value,8.000000,℃,1778314683,0,1,1,2,2;143,Return air temperature alarm value,35.000000,℃,1778314683,0,1,1,2,2;144,Return wind high humidity alarm value,95.000000,%,1778314683,0,1,1,2,2;145,Return air low warning value,8.000000,%,1778314683,0,1,1,2,2;146,Airflow loss temperature alarm value,16.000000,℃,1778314683,0,1,1,2,2;147,Compressor control mode,SupplyAir[0],,1778314683,0,1,1,0,5;148,Fan control mode,SupplyAir[0],,1778314683,0,1,1,0,5;149,Fan remote �� T,2.000000,℃,1778314683,0,1,1,2,2;150,Fan return air �� T,12.000000,℃,1778314683,0,1,1,2,2;151,Compressor temperature proportional band,5.000000,℃,1778314683,0,1,1,2,2;152,Compressor temperature integration time,300.000000,Sec,1778314683,0,1,1,2,2;153,Compressor temperature differential time,0.000000,Sec,1778314683,0,1,1,2,2;154,Fan temperature proportional band,13.000000,℃,1778314683,0,1,1,2,2;155,Fan temperature integration time,240.000000,Sec,1778314683,0,1,1,2,2;156,Fan temperature differential time,0.000000,Sec,1778314683,0,1,1,2,2;6,Model selection,25KW[1],,1778314683,0,1,1,0,5;158,Compressor start demand,50.000000,%,1778314683,0,1,1,2,2;159,Compressor stops demand,-150.000000,%,1778314683,0,1,1,2,2;160,Compressor minimum capacity,30.000000,%,1778314683,0,1,1,2,2;316,Compressor standard capacity,100.000000,%,1778314683,0,1,1,2,2;162,Compressor maximum capacity,125.000000,%,1778314683,0,1,1,2,2;163,Compressor maximum capacity,110.000000,%,1778314683,0,1,1,2,2;164,Compressor start capacity,40.000000,%,1778314683,0,1,1,2,2;165,Compressor dehumidification capacity increases,15.000000,%,1778314683,0,1,1,2,2;166,Maximum capacity of the compressor running time,120.000000,Min,1778314683,0,1,1,2,2;167,Compressor start time,180.000000,Sec,1778314683,0,1,1,2,2;168,Compressor output dead zone,2.800000,%,1778314683,0,1,1,2,2;169,Oil return cycle,240.000000,Min,1778314683,0,1,1,2,2;170,Oil return running time,5.000000,Min,1778314683,0,1,1,2,2;171,Oil return capacity,60.000000,%,1778314683,0,1,1,2,2;172,Fan minimum speed,40.000000,%,1778314683,0,1,1,2,2;173,Fan standard speed,75.000000,%,1778314683,0,1,1,2,2;174,Fan minimum speed CFC,0.000000,%,1778314683,0,1,1,2,2;175,Fan standard speed CFC,100.000000,%,1778314683,0,1,1,2,2;176,Fan humidification speed,75.000000,%,1778314683,0,1,1,2,2;177,Fan analog output lower limit,30.000000,%,1778314683,0,1,1,2,2;178,Fan analog output upper limit,100.000000,%,1778314683,0,1,1,2,2;179,Fan low speed step,0.100000,%/s,1778314683,0,1,1,2,2;180,Fan high speed step,1.000000,%/s,1778314683,0,1,1,2,2;181,Fan down delay,5.000000,Sec,1778314683,0,1,1,2,2;182,EEV time constant,60.000000,Sec,1778314683,0,1,1,2,2;183,EEV superheat setting,6.000000,℃,1778314683,0,1,1,2,2;184,EEV MOP pressure limit,11.000000,Bar,1778314683,0,1,1,2,2;185,EEV start opening degree,65.000000,%,1778314683,0,1,1,2,2;186,The EEV valve closes the superheat,6.000000,℃,1778314683,0,1,1,2,2;187,Electric heating fault alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;188,Exhaust temperature sensor fault alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;189,Fan failure alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;190,EEV communication fault alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;191,The refrigerant alarm attribute is not selected,TurnON[1],,1778314683,0,1,1,0,5;192,Insufficient refrigerant alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;193,Inhalation temperature sensor fault alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;194,Low pressure sensor failure alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;195,Compressor Drive Communication Fault Alarm Attribute,TurnON[1],,1778314683,0,1,1,0,5;196,Compressor drive failure failure alarm,TurnON[1],,1778314683,0,1,1,0,5;197,Compressor radiator over temperature alarm,TurnON[1],,1778314683,0,1,1,0,5;198,Compressor overcurrent alarm,TurnON[1],,1778314683,0,1,1,0,5;199,Compressor phase failure protection alarm,TurnON[1],,1778314683,0,1,1,0,5;200,Busbar voltage exception alarm,TurnON[1],,1778314683,0,1,1,0,5;201,Humidifier fault alarm attribute,TurnON[2],,1778314683,0,1,1,0,5;202,Group-controlled address repeat alarm attribute,TurnON[1],,1778314683,0,1,1,0,5;203,EEV start time,60.000000,Sec,1778314683,0,1,1,2,2;204,Fan maximum speed low pressure point,5.800000,Bar,1778314683,0,1,1,2,2;205,Fan speed low point,7.500000,Bar,1778314683,0,1,1,2,2;206,Fan down the low pressure point,11.500000,Bar,1778314683,0,1,1,2,2;207,Minimum speed of the fan,13.500000,Bar,1778314683,0,1,1,2,2;208,Compressor minimum output low pressure point,5.800000,Bar,1778314683,0,1,1,2,2;209,Compressor capacity reduces low pressure point,7.500000,Bar,1778314683,0,1,1,2,2;210,Compressor capacity increases low pressure point,11.500000,Bar,1778314683,0,1,1,2,2;211,Compressor maximum output low pressure point,13.500000,Bar,1778314683,0,1,1,2,2;212,Remote temperature sense 1 correction value,0.000000,℃,1778314683,0,1,1,2,2;213,Remote temperature 2 correction value,0.000000,℃,1778314683,0,1,1,2,2;214,Remote temperature 3 correction value,0.000000,℃,1778314683,0,1,1,2,2;215,Remote temperature sense 4 correction value,0.000000,℃,1778314683,0,1,1,2,2;216,Remote temperature 5 correction value,0.000000,℃,1778314683,0,1,1,2,2;217,Remote temperature 6 correction value,0.000000,℃,1778314683,0,1,1,2,2;218,Remote temperature 7 correction value,0.000000,℃,1778314683,0,1,1,2,2;219,Remote temperature 8 correction value,0.000000,℃,1778314683,0,1,1,2,2;220,Remote temperature 9 correction value,0.000000,℃,1778314683,0,1,1,2,2;221,Remote temperature 10 correction value,0.000000,℃,1778314683,0,1,1,2,2;222,Fan failure mode,CoolingOnly[0],,1778314683,0,1,1,0,5;223,Power overrun alarm value,12.000000,%,1778314683,0,1,1,2,2;224,Power undervoltage alarm value,-14.000000,%,1778314683,0,1,1,2,2;225,25 Model Resonance Point 1,0.000000,%,1778314683,0,1,1,2,2;226,25 Model Resonance Point 1 Bandwidth,0.500000,%,1778314683,0,1,1,2,2;227,25 Model Resonance Point 2,0.000000,%,1778314683,0,1,1,2,2;228,25 Model Resonance Point 2 Bandwidth,0.500000,%,1778314683,0,1,1,2,2;229,25 Model Resonance Point 3,0.000000,%,1778314683,0,1,1,2,2;230,25 models resonant point 3 bandwidth,0.500000,%,1778314683,0,1,1,2,2;231,25 Model Resonance Point 4,0.000000,%,1778314683,0,1,1,2,2;232,25 Model Resonance Point 4 Bandwidth,0.500000,%,1778314683,0,1,1,2,2;233,25 Model Resonance Point 5,0.000000,%,1778314683,0,1,1,2,2;234,25 Model Resonance Point 5 Bandwidth,0.500000,%,1778314683,0,1,1,2,2;235,35 Model Resonance Point 1,0.000000,%,1778314683,0,1,1,2,2;236,35 models resonant point 1 bandwidth,0.500000,%,1778314683,0,1,1,2,2;237,35 Model Resonance Point 2,0.000000,%,1778314683,0,1,1,2,2;238,35 models resonant point 2 bandwidth,0.500000,%,1778314683,0,1,1,2,2;239,35 Model Resonance Point 3,0.000000,%,1778314683,0,1,1,2,2;240,35 models resonant point 3 bandwidth,0.500000,%,1778314683,0,1,1,2,2;241,35 Model Resonance Point 4,0.000000,%,1778314683,0,1,1,2,2;242,35 models resonant point 4 bandwidth,0.500000,%,1778314683,0,1,1,2,2;243,35 Model Resonance Point 5,0.000000,%,1778314683,0,1,1,2,2;244,35 models resonant point 5 bandwidth,0.500000,%,1778314683,0,1,1,2,2;252,High pressure alarm switch,Open[1],,1778314683,0,1,1,0,5;253,High water level alarm switch,Open[1],,1778314683,0,1,1,0,5;254,Floor overflow alarm switch,Open[1],,1778314683,0,1,1,0,5;255,Electric heating fault alarm switch,Open[1],,1778314683,0,1,1,0,5;256,Filter plugging alarm switch,Open[1],,1778314683,0,1,1,0,5;257,Humidifier fault alarm switch,Open[1],,1778314683,0,1,1,0,5;258,Water level switch,Close[0],,1778314683,0,1,1,0,5;259,Remote switch,Close[0],,1778314683,0,1,1,0,5;260,Condensate pump output,OutputClose[0],,1778314683,0,1,1,0,5;261,Compressor output,OutputMove[1],,1778314683,0,1,1,0,5;262,Electric heating output,OutputClose[0],,1778314683,0,1,1,0,5;7,Fan output,OutputMove[0],,1778314683,0,1,1,0,5;264,Humidifier output,OutputClose[0],,1778314683,0,1,1,0,5;265,Public alarm output,OutputMove[1],,1778314683,0,1,1,0,5;266,Liquid circuit solenoid valve output,OutputMove[1],,1778314683,0,1,1,0,5;267,High pressure alarm,OutputClose[0],,1778314683,0,1,1,0,5;300,Manually clear the start and stop records,No[0],,1778314683,0,1,1,0,5;301,Manually clear the history alarm,No[0],,1778314683,0,1,1,0,5;302,Manual run allowed,No[0],,1778314683,0,1,1,0,5;303,reset,No[0],,1778314683,0,1,1,0,5;304,Vacuum state,No[0],,1778314683,0,1,1,0,5;305,Humidification function enabled,Disable[0],,1778314683,0,1,1,0,5;306,Dehumidification function enabled,Enable[1],,1778314683,0,1,1,0,5;307,Clear high pressure exception,No[0],,1778314683,0,1,1,0,5;308,Dehumidification lock release,No[0],,1778314683,0,1,1,0,5;309,Monitor shutdown enable,Enable[1],,1778314683,0,1,1,0,5;310,Soft shutdown status,No[0],,1778314683,0,1,1,0,5;311,New alarm flag,GenerateNewAlarm[1],,1778314683,0,1,1,0,5;312,Panel shutdown flag,NotIn[0],,1778314683,0,1,1,0,5;313,Monitor the shutdown flag,NotIn[0],,1778314683,0,1,1,0,5;314,Remote shutdown flag,NotIn[0],,1778314683,0,1,1,0,5;315,Filter maintenance,No[0],,1778314683,0,1,1,0,5;161,Fan 1 state,Normal[0],,1778314683,0,1,1,0,5;317,Fan 2 state,Normal[0],,1778314683,0,1,1,0,5;318,Fan 3 state,Normal[0],,1778314683,0,1,1,0,5;319,Fan 4 state,Normal[0],,1778314683,0,1,1,0,5;320,Remote switch machine input polarity,NC[0],,1778314683,0,1,1,0,5;321,Common alarm output polarity,NO[1],,1778314683,0,1,1,0,5;322,Switch machine switch,No[0],,1778314683,0,1,1,0,5;8,Refrigeration flag,In[0],,1778314683,0,1,1,0,5;11,Heating flag,NotIn[1],,1778314683,0,1,1,0,5;12,Humidification mark,NotIn[1],,1778314683,0,1,1,0,5;13,Dehumidification mark,NotIn[1],,1778314683,0,1,1,0,5;342,Service Message Alert Enable,TurnON[1],,1778314683,0,1,1,0,5;343,Manual wheel patrol,No[0],,1778314683,0,1,1,0,5;344,Cascading function,Disable[0],,1778314683,0,1,1,0,5;346,Power supply system,50HZ[0],,1778314683,0,1,1,0,5;91,Communication Status,Normal[0],,1778314683,0,1,1,0,5;



ENV_THD
https://vertiv.cn-sh-1.daocloud.io/cgi-bin/p05_equip_sample.cgi?sand=0.22718244959718248&_equipId=-98&_op_type=1

5005,ENV_THD,THD_SENSOR^3,RACK1 Cool Aisle Top Temp,21.700001,℃,1778314638,0,1,1,2,2;4,RACK1 Cool Aisle Middle Temp,22.900000,℃,1778314638,0,1,1,2,2;5,RACK1 Cool Aisle Hum,40.400002,%,1778314638,0,1,1,2,2;6,RACK1 Cool Aisle Bottom Temp,19.400000,℃,1778314638,0,1,1,2,2;7,RACK1 Hot Aisle Top Temp,29.400000,℃,1778314638,0,1,1,2,2;8,RACK1 Hot Aisle Middle Temp,31.400000,℃,1778314638,0,1,1,2,2;9,RACK1 Hot Aisle Hum,24.400000,%,1778314638,0,1,1,2,2;10,RACK1 Hot Aisle Bottom Temp,26.400000,℃,1778314638,0,1,1,2,2;12,RACK1 Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;16,RACK1 Hot Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;22,RACK2 Cool Aisle Top Temp,24.500000,℃,1778314638,0,1,1,2,2;23,RACK2 Cool Aisle Middle Temp,23.500000,℃,1778314638,0,1,1,2,2;24,RACK2 Cool Aisle Hum,42.000000,%,1778314638,0,1,1,2,2;25,RACK2 Cool Aisle Bottom Temp,20.900000,℃,1778314638,0,1,1,2,2;26,RACK2 Hot Aisle Top Temp,28.900000,℃,1778314638,0,1,1,2,2;27,RACK2 Hot Aisle Middle Temp,31.400000,℃,1778314638,0,1,1,2,2;28,RACK2 Hot Aisle Hum,27.299999,%,1778314638,0,1,1,2,2;29,RACK2 Hot Aisle Bottom Temp,29.400000,℃,1778314638,0,1,1,2,2;31,RACK2 Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;35,RACK2 Hot Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;41,RACK3 Cool Aisle Top Temp,26.200001,℃,1778314638,0,1,1,2,2;42,RACK3 Cool Aisle Middle Temp,23.400000,℃,1778314638,0,1,1,2,2;43,RACK3 Cool Aisle Hum,41.500000,%,1778314638,0,1,1,2,2;44,RACK3 Cool Aisle Bottom Temp,24.000000,℃,1778314638,0,1,1,2,2;45,RACK3 Hot Aisle Top Temp,28.299999,℃,1778314638,0,1,1,2,2;46,RACK3 Hot Aisle Middle Temp,27.200001,℃,1778314638,0,1,1,2,2;47,RACK3 Hot Aisle Hum,32.700001,%,1778314638,0,1,1,2,2;48,RACK3 Hot Aisle Bottom Temp,26.400000,℃,1778314638,0,1,1,2,2;50,RACK3 Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;54,RACK3 Hot Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;60,RACK4 Cool Aisle Top Temp,23.600000,℃,1778314638,0,1,1,2,2;61,RACK4 Cool Aisle Middle Temp,21.799999,℃,1778314638,0,1,1,2,2;62,RACK4 Cool Aisle Hum,46.299999,%,1778314638,0,1,1,2,2;63,RACK4 Cool Aisle Bottom Temp,22.100000,℃,1778314638,0,1,1,2,2;64,RACK4 Hot Aisle Top Temp,29.500000,℃,1778314638,0,1,1,2,2;65,RACK4 Hot Aisle Middle Temp,27.000000,℃,1778314638,0,1,1,2,2;66,RACK4 Hot Aisle Hum,33.099998,%,1778314638,0,1,1,2,2;67,RACK4 Hot Aisle Bottom Temp,27.799999,℃,1778314638,0,1,1,2,2;69,RACK4 Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;73,RACK4 Hot Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;79,RACK5 Cool Aisle Top Temp,20.700001,℃,1778314638,0,1,1,2,2;80,RACK5 Cool Aisle Middle Temp,20.799999,℃,1778314638,0,1,1,2,2;81,RACK5 Cool Aisle Hum,50.099998,%,1778314638,0,1,1,2,2;82,RACK5 Cool Aisle Bottom Temp,19.600000,℃,1778314638,0,1,1,2,2;83,RACK5 Hot Aisle Top Temp,26.100000,℃,1778314638,0,1,1,2,2;84,RACK5 Hot Aisle Middle Temp,24.600000,℃,1778314638,0,1,1,2,2;85,RACK5 Hot Aisle Hum,38.099998,%,1778314638,0,1,1,2,2;86,RACK5 Hot Aisle Bottom Temp,26.200001,℃,1778314638,0,1,1,2,2;88,RACK5 Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;92,RACK5 Hot Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;155,RACK PMC Cool Aisle Top Temp,21.700001,℃,1778314638,0,1,1,2,2;156,RACK PMC Cool Aisle Middle Temp,22.400000,℃,1778314638,0,1,1,2,2;157,RACK PMC Cool Aisle Hum,44.799999,%,1778314638,0,1,1,2,2;158,RACK PMC Cool Aisle Bottom Temp,19.000000,℃,1778314638,0,1,1,2,2;159,RACK PMC Hot Aisle Top Temp,26.200001,℃,1778314638,0,1,1,2,2;160,RACK PMC Hot Aisle Middle Temp,25.200001,℃,1778314638,0,1,1,2,2;161,RACK PMC Hot Aisle Hum,37.599998,%,1778314638,0,1,1,2,2;162,RACK PMC Hot Aisle Bottom Temp,24.799999,℃,1778314638,0,1,1,2,2;164,RACK PMC Cool Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;168,RACK PMC Hot Aisle Door Status,Normal[0],,1778314638,0,1,1,0,5;3010,RACK1 THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;3011,RACK2 THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;3012,RACK3 THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;3013,RACK4 THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;3014,RACK5 THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;3018,RACK PMC THD Comm Status,Normal[0],,1778314638,0,1,1,0,5;10000,RACK1,1,,1778314638,0,1,1,0,3;10001,RACK2,2,,1778314638,0,1,1,0,3;10002,RACK3,3,,1778314638,0,1,1,0,3;10003,RACK4,4,,1778314638,0,1,1,0,3;10004,RACK5,5,,1778314638,0,1,1,0,3;10008,RACK PMC,9,,1778314638,0,1,1,0,3;10012,High Temp Alarm Rack Count,0,,1778314638,0,1,1,0,3;



UPS
https://vertiv.cn-sh-1.daocloud.io/cgi-bin/p05_equip_sample.cgi?sand=0.3364588858487123&_equipId=26&_op_type=1

491,UPS_1,ENP_UPS_ITA2[COM]^2,Phase A Input Voltage,217.800003,V,1778314661,0,1,1,2,2;3,Phase B Input Voltage,219.600006,V,1778314661,0,1,1,2,2;4,Phase C Input Voltage,221.899994,V,1778314661,0,1,1,2,2;5,Phase A Output Voltage,219.699997,V,1778314661,0,1,1,2,2;6,Phase B Output Voltage,221.199997,V,1778314661,0,1,1,2,2;7,Phase C Output Voltage,218.699997,V,1778314661,0,1,1,2,2;8,Phase A Output Current,15.200000,A,1778314661,0,1,1,2,2;9,Phase B Output Current,2.800000,A,1778314661,0,1,1,2,2;10,Phase C Output Current,15.000000,A,1778314661,0,1,1,2,2;11,Output Frequency,49.970001,HZ,1778314661,0,1,1,2,2;49,Input Phase Number,Three Phase[3],,1778314661,0,1,1,1,5;16,Line Ab Input Voltage,378.399994,V,1778314661,0,1,1,2,2;69,Line Bc Input Voltage,381.299988,V,1778314661,0,1,1,2,2;63,Line Ca Input Voltage,381.799988,V,1778314661,0,1,1,2,2;124,Phase A Input Current,10.300000,A,1778314661,0,1,1,2,2;125,Phase B Input Current,9.900000,A,1778314661,0,1,1,2,2;126,Phase C Input Current,10.400000,A,1778314661,0,1,1,2,2;12,System Input Frequency,50.000000,HZ,1778314661,0,1,1,2,2;23,Phase A Input Power Factor,0.990000,,1778314661,0,1,1,2,2;71,Phase B Input Power Factor,0.990000,,1778314661,0,1,1,2,2;77,Phase C Input Power Factor,0.990000,,1778314661,0,1,1,2,2;143,Bypass A Voltage A,221.500000,V,1778314661,0,1,1,2,2;144,Bypass B Voltage B,220.000000,V,1778314661,0,1,1,2,2;145,Bypass C Voltage C,223.399994,V,1778314661,0,1,1,2,2;29,Bypass Line Ab Voltage,382.200012,V,1778314661,0,1,1,2,2;30,Bypass Line Bc Voltage,384.299988,V,1778314661,0,1,1,2,2;31,Bypass Line Ca Voltage,386.200012,V,1778314661,0,1,1,2,2;146,Bypass Frequency,49.970001,HZ,1778314661,0,1,1,2,2;34,Output Phase Number,Three Phase[3],,1778314661,0,1,1,1,5;35,Phase A Output Power Factor,0.980000,,1778314661,0,1,1,2,2;36,Phase B Output Power Factor,0.940000,,1778314661,0,1,1,2,2;37,Phase C Output Power Factor,0.990000,,1778314661,0,1,1,2,2;134,Local Phase A Output Active Power,3.110000,KW,1778314661,0,1,1,2,2;135,Local Phase B Output Active Power,0.440000,KW,1778314661,0,1,1,2,2;136,Local Phase C Output Active Power,3.040000,KW,1778314661,0,1,1,2,2;44,Local Phase A Output Apparent Power,3.150000,KVA,1778314661,0,1,1,2,2;45,Local Phase B Output Apparent Power,0.490000,KVA,1778314661,0,1,1,2,2;46,Local Phase C Output Apparent Power,3.060000,KVA,1778314661,0,1,1,2,2;13,Local Phase A Output Load Percen,47.200001,%,1778314661,0,1,1,2,2;14,Local Phase B Output Load Percen,7.400000,%,1778314661,0,1,1,2,2;15,Local Phase C Output Load Percen,45.799999,%,1778314661,0,1,1,2,2;54,System Phase A Output Active Power,3.130000,KW,1778314661,0,1,1,2,2;55,System Phase B Output Active Power,0.430000,KW,1778314661,0,1,1,2,2;56,System Phase C Output Active Power,3.030000,KW,1778314661,0,1,1,2,2;57,System Phase A Output Apparent Power,3.170000,KVA,1778314661,0,1,1,2,2;58,System Phase B Output Apparent Power,0.490000,KVA,1778314661,0,1,1,2,2;59,System Phase C Output Apparent Power,3.050000,KVA,1778314661,0,1,1,2,2;60,Parallel Machine Number,1.000000,,1778314661,0,1,1,2,2;62,Ups Running Time,2413.000000,day,1778314661,0,1,1,2,2;18,Battery Voltage,219.100006,V,1778314661,0,1,1,2,2;64,Battery Charging Current,0.200000,A,1778314661,0,1,1,2,2;65,Battery Discharge Current,0.000000,A,1778314661,0,1,1,2,2;66,Negative Battery Voltage,218.600006,V,1778314661,0,1,1,2,2;67,Negative Battery Charge Current,0.070000,A,1778314661,0,1,1,2,2;68,Negative Battery Discharge Current,0.000000,A,1778314661,0,1,1,2,2;17,Battery Backup Time,33.900002,Min,1778314661,0,1,1,2,2;24,Ambient Temperature,20.500000,℃,1778314661,0,1,1,2,2;72,Battery Current Capacity,100.000000,%,1778314661,0,1,1,2,2;73,Battery Discharge Times,59.000000,,1778314661,0,1,1,2,2;75,Input Power,261917.000000,KWH,1778314661,0,1,1,2,2;76,Output Power,251038.000000,KWH,1778314661,0,1,1,2,2;25,Power Supply,Utility Online[1],,1778314661,0,1,1,0,5;167,Battery Discharging Time,3054.000000,s,1778314661,0,1,1,2,2;79,Input Power Status,Utility Online[0],,1778314661,0,1,1,0,5;27,Battery Status,Float Charging[1],,1778314661,0,1,1,0,5;81,Battery Negative Group Status,Float Charging[1],,1778314661,0,1,1,0,5;82,Charger Status,Charger On[0],,1778314661,0,1,1,0,5;83,Parallel System Power State,Main Inverter Power Supply[0],,1778314661,0,1,1,0,5;84,Ineer Network Connection Status,Disconnected[1],,1778314661,0,1,1,0,5;456,Communication Status,Normal[0],,1778314661,0,1,1,0,5;