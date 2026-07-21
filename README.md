# Vertiv RDU-A G2智能监控单元

<img width="1919" height="1057" alt="iShot_2026-05-11_08 50 17" src="https://github.com/user-attachments/assets/04faf042-81a2-495f-9847-2b908a624b58" />

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
config.example.yaml         Example exporter configuration
vertiv_grafana_dashboard.json  Importable Grafana dashboard
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
    host: "https://vertiv.example.local"
    username: "admin"
    password: "plain_password"
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
- `target.name`: value used as the Prometheus `instance` label
- `host`: Vertiv web base URL
- `username` / `password`: plain-text login values; the exporter encodes them automatically before calling `login.cgi`
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

- The YAML file contains plain-text credentials. Keep `config.yaml` out of version control and restrict it to the exporter user, for example with `chmod 600 config.yaml`.
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

## Run

Use the example config as a starting point:

```bash
cp config.example.yaml config.yaml
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

## Docker

Local single-platform build:

```bash
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t vertiv-exporter:1.0.0 .
```

Multi-platform build and push:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t your-registry/vertiv-exporter:1.0.0 \
  --push .
```

Run it with a mounted config file:

```bash
docker run --rm -p 9101:9101 \
  --user "$(id -u):$(id -g)" \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  vertiv-exporter:1.0.0
```

The image expects the config file at `/app/config.yaml`. The local example runs with the host UID/GID so a `chmod 600` bind-mounted config remains readable without making it world-readable. Secret or ConfigMap mounts can keep the image's default `nonroot` user when their permissions allow UID `65532` to read the file. If `exporter.metrics_file` is set, mount that override file separately as read-only and use its in-container path in `config.yaml`.

The build context is restricted by `.dockerignore` to the build definition, `go.mod`, `go.sum`, and production Go source files. Local credentials, Git history, dashboards, documentation, tests, caches, and build artifacts are not sent to the Docker builder.

## Grafana Dashboard

`vertiv_grafana_dashboard.json` uses the Grafana dashboard v2 resource schema. Its Prometheus queries reference a `datasource` variable instead of an environment-specific data source UID, and its instance/device selections start empty. Select the target Prometheus data source after importing the dashboard.

The checked-in resource intentionally omits server-managed UID, resource version, namespace, user, and folder metadata. Add the target namespace or folder annotation in the deployment environment when required.

## Supported Metric Groups

Every device metric includes `instance`, `device`, and `equip_id`. THD and UPS metrics add the labels noted below. Exporter health is reported through:

- `vertiv_exporter_up{instance}`: `1` when every configured device for the target was collected, otherwise `0`
- `vertiv_exporter_scrape_duration_seconds`: histogram of complete collector run duration
- `vertiv_exporter_scrape_failures_total`: total device fetch failures

### AC

- Temperature
- Humidity
- Pressure
- Electrical values
- Compressor, fan, EEV, runtime, alarm attributes, system status

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
