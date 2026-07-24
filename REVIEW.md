# Vertiv Exporter 代码审核报告

审核范围：`cmd/vertiv_exporter`、`internal/client`、`internal/collector`、`internal/config` 以及默认指标定义。审核基于当前工作区代码，不包含真实 Vertiv 设备的压力测试结果。

## 结论

- **P0：0 项。** 未发现已证实的凭证硬编码、无限期 HTTP 阻塞或单个设备请求失败直接令 `/metrics` 返回 500 的问题。
- **P1：10 项。** 主要集中在会话失效识别、并发上限、单设备失败隔离、自定义指标校验、Label 语义和指标类型。
- **P2：6 项。** 主要是超时细化、Prometheus 基础单位、日志和 HTTP 服务防护。

当前实现已有以下有效保护：

- 每个 target 只创建一个带 `cookiejar` 的 HTTP client，会话 Cookie 可跨登录、keepalive 和采集请求复用（`internal/client/client.go:28-58`、`internal/collector/collector.go:53-63`）。
- 设备返回 `401` 或 `302` 时会自动登录并重试一次（`internal/client/client.go:144-151`）。
- 单次 collector 执行使用可配置的整体超时，默认 10 秒（`internal/config/config.go:58-60`、`internal/collector/collector.go:147-151`）。
- 不同 target 当前并行采集；同一 target 的设备顺序采集，且 client mutex 会串行化该 target 的 CGI 请求（`internal/collector/collector.go:159-185`、`internal/client/client.go:120-125`）。
- 设备请求失败会记录日志、增加失败 Counter 并将 target 的 `vertiv_exporter_up` 置为 0；collector 本身不返回错误，因此常规失败不会直接让 `/metrics` 变成 500（`internal/collector/collector.go:164-195`）。

## P0（必须修）

当前未发现 P0 问题。

## P1（建议优先修）

| 位置 | 维度 | 问题描述 | 修改建议 |
| --- | --- | --- | --- |
| `internal/client/client.go:84-89`、`internal/client/client.go:144-151` | 会话健壮性 | 登录函数把所有 `3xx` 都视为成功，但普通请求只把 `302` 和 `401` 识别为会话失效。设备若用 `301/303/307/308` 跳回登录页，或用 `200` 返回登录 HTML，代码不会自动重登，最终只表现为解析失败。 | 统一实现 `isAuthExpired(resp, body)`：检查所有登录重定向、`Location`、内容类型和已知登录页标记；最多自动重登一次，避免循环。登录成功还应验证预期 Cookie 或响应标记，而不是仅按状态码判断。 |
| `internal/collector/collector.go:147-187`、`internal/client/client.go:120-125` | 性能 / 并发 | 每个 scrape 为所有 target 各启动一个 goroutine，没有全局并发上限；并发 `/metrics` 请求还会重复启动整轮采集。同一 target 虽被 mutex 串行化，但重叠 scrape 会在锁后排队并占用 goroutine，target 数量较大时也可能同时打满设备。 | 增加 `exporter.max_concurrent_targets`，使用 `errgroup.SetLimit` 或 semaphore；再用 scrape mutex/singleflight 合并重叠采集，或明确拒绝/复用正在进行的采集结果。上线前用目标规模做延迟与连接数压测。 |
| `internal/collector/collector.go:164-178` | 故障隔离 | 一个设备失败后立即 `break`，同 target 后续设备完全不采集。虽然 `/metrics` 不会 500，但会丢失本可用设备的数据，且只有 target 级 `up`，无法判断具体失败设备。 | 失败后 `continue`，保留其他设备采集；新增 `vertiv_exporter_device_up{target,device,equip_id}`，target `up` 取所有设备结果的聚合。 |
| `internal/collector/metadata.go:13-64`、`internal/collector/collector.go:65-68`、`internal/collector/collector.go:240-247` | 健壮性 | 自定义 metrics Markdown 仅用正则取出任意名称，没有验证 Prometheus metric name、重复名称或冲突 help。无效 `Desc` 最迟可能在 `MustNewConstMetric` 路径触发 panic/Gather 错误，使 `/metrics` 异常。 | 加载时一次性校验名称合法性、Field ID 唯一、metric name 唯一及 help 一致性；构建所有 `Desc` 后立即检查错误并让进程启动失败，避免把配置错误推迟到 scrape。 |
| `internal/collector/metadata.go:21-32` | 配置健壮性 | 用户显式配置 `metrics_file` 后，如果文件不存在或无权限，代码静默回退到内置映射。部署拼写错误会产生“服务正常但指标映射错误”的隐蔽故障。 | 只有空路径才使用默认映射；非空路径打开失败应返回带路径的错误。若确需回退，增加显式 `metrics_file_optional` 开关并记录 warning。 |
| `internal/config/config.go:26-32` | 凭证安全 | 账号密码未硬编码在源码，但只支持 YAML 明文值；配置泄漏、进程诊断包或错误的 ConfigMap 挂载都会暴露凭证。 | 支持 `username_file` / `password_file` 或 `${ENV_VAR}` 引用，并优先推荐 Secret 文件挂载；对冲突配置做校验。继续保持 `config.yaml` 不入库和最小文件权限。 |
| `internal/client/parser.go:37-55`、`internal/client/parser.go:60-94` | 数据完整性 | 解析器静默跳过所有不合法 record，只要至少有一个字段成功就返回成功。设备协议变化或局部截断会造成大量指标悄然消失，但 `vertiv_exporter_up` 仍可能为 1。 | 返回“成功样本 + 跳过数量/原因”；设置可配置的最低成功率或关键字段检查。至少导出 `vertiv_exporter_parse_failures_total{target,device}` 并记录受限采样日志。 |
| `internal/collector/collector.go:22`、`internal/collector/collector.go:78-82` | Label 规范 | Exporter 用 `instance` 表示配置中的 Vertiv target，但 Prometheus 也默认用 `instance` 表示 scrape endpoint。在默认 `honor_labels: false` 下，设备侧值会被改名为 `exported_instance`，README 语义、查询和多 target 识别容易混淆。 | 新版本将设备目标 Label 改为 `target` 或 `vertiv_target`，保留一个发布周期的兼容指标/recording rule；短期若继续使用 `instance`，必须在 Prometheus 示例中明确 `honor_labels` 策略及其副作用。 |
| `internal/collector/default_metrics.go:85-87`、`internal/collector/default_metrics.go:101-112`、`internal/collector/collector.go:240-247` | 指标类型 | 多个 AC 累计指标以 `_total` 结尾，但统一通过 `prometheus.GaugeValue` 导出。名称表达 Counter，TYPE 却是 Gauge，`rate()`、重置处理和监控工具推断都会产生歧义。 | 核实设备值是否单调累计；若是，按 Field ID 元数据增加 ValueType 并导出 Counter。若设备值可下降或重启后语义不稳定，应移除 `_total` 并通过兼容期完成重命名。 |
| `internal/collector/collector.go:104-110`、`internal/collector/collector.go:216-232` | Label 基数 | `vertiv_ac_signal_value` 把设备返回的任意 `signal_name` 直接作为 Label；固件、语言、空格或名称变化会持续创建新时间序列。`occurrence` 还会在同名字段集合变化时产生额外 churn。 | 默认关闭 raw metric 或增加 allowlist；导出 `signal_id` 等稳定低基数标识，名称放入受控 info 映射。至少记录/限制每设备信号数，并在容量评估中计算 `target × device × signal` 序列规模。 |

## P2（可选优化）

| 位置 | 维度 | 问题描述 | 修改建议 |
| --- | --- | --- | --- |
| `cmd/vertiv_exporter/main.go:69-70`、`internal/collector/collector.go:147-187` | 响应延迟 | `/metrics` 采用同步按需采集，所以响应时间直接等于最慢 target 的 CGI 链路，直到整体 scrape timeout。这是 exporter 常见模型，但慢设备会直接消耗 Prometheus scrape 预算。 | 保持按需模型时，让 Prometheus `scrape_timeout` 明确大于 exporter timeout 并监控 p95；如果设备长期较慢，再评估后台定时采集 + 缓存快照，并公开数据时间戳/陈旧状态。 |
| `internal/client/client.go:48-57`、`internal/collector/collector.go:53-63` | 超时 / 启动性能 | HTTP client timeout 固定 15 秒；初始登录在 target 循环中串行执行，并使用进程 context，而不是 `scrape_timeout`。大量离线 target 会显著拖慢启动。 | 增加 `request_timeout` / `login_timeout` 配置；初始登录使用受限并发和独立 deadline。保持 request timeout 不大于整体 scrape timeout，或清晰说明两者优先级。 |
| `internal/client/client.go:159-164` | 内存健壮性 | 成功响应使用无上限 `io.ReadAll`。异常设备若返回超大响应会造成不必要的内存峰值。 | 使用 `io.LimitReader` 和可配置的合理上限，超过上限返回明确错误并增加失败指标。 |
| `internal/collector/default_metrics.go:5-185`、`internal/collector/ups.go:114-153` | 命名 / 单位 | 多数名称带单位，但仍使用 `bar`、`minutes`、`hours`、`days`、`kWh`、`kVA`、`hz` 等非 Prometheus 推荐基础单位；另有 `_seconds` 与 `_sec` 并存。直接重命名会破坏现有查询。 | 新主版本逐步迁移为基础单位（如 seconds、volts、amperes、joules、hertz）并在采集时换算；提供 recording rules 和弃用窗口。至少先统一新指标的 suffix 规则。 |
| `internal/collector/collector.go:89-92`、`internal/collector/collector.go:168-170`、`internal/collector/collector.go:262-264` | 可观测性 / 日志 | 失败 Counter 没有 target、device 或阶段 Label；日志使用全局 `log.Printf`，没有级别、结构化字段或稳定错误类别。线上只能依赖文本搜索定位。 | 采用可注入的结构化 logger，统一 `target/device/equip_id/stage/error_kind` 字段；为失败 Counter 增加受控低基数的 `target`、`stage`，设备级状态由 `device_up` 表达。 |
| `cmd/vertiv_exporter/main.go:80-84` | HTTP 健壮性 | Server 仅配置 `ReadHeaderTimeout`，没有 `IdleTimeout` 和连接级防护；`WriteTimeout` 需结合最大 scrape 时间谨慎设置。 | 增加合理的 `IdleTimeout`；如设置 `WriteTimeout`，应大于 exporter scrape timeout 并留出编码响应余量。限制不必要的并发连接，避免慢客户端长期占用资源。 |

## 建议实施顺序

1. 修复认证失效识别、自定义指标校验和显式 metrics file 错误。
2. 增加 target 并发上限、重叠 scrape 保护和设备级故障隔离。
3. 引入安全的凭证来源及设备级可观测性。
4. 设计 `instance` → `target` 与 AC `_total` 类型修正的兼容迁移。
5. 在主版本窗口统一基础单位和剩余命名。
