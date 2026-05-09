# Vertiv RDU-A G2智能监控单元

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
cmd/vertiv_exporter/        Program entrypoint
internal/client/            Login, keepalive, CGI fetching, response parsing
internal/collector/         Prometheus collectors and device-specific mappings
internal/config/            YAML config loader
docs/vertiv_exporter_design.md  Design notes and captured response examples
docs/vertiv_ac_metrics_list.md  AC metric reference
docs/vertiv_thd_metrics_list.md THD metric reference
docs/vertiv_ups_metrics_list.md UPS metric reference
```

## Configuration

Example config:

```yaml
exporter:
  listen_address: ":9101"
  metrics_path: "/metrics"
  scrape_timeout: 10s
  debug_response: false

targets:
  - name: "dc-rack-01"
    host: "https://vertiv.example.local"
    username: "encoded_username"
    password: "encoded_password"
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
        equip_id: 491
```

### Config Fields

- `target.name`: value used as the Prometheus `instance` label
- `host`: Vertiv web base URL
- `username` / `password`: plain-text login values; the exporter encodes them automatically before calling `login.cgi`
- `tls_skip_verify`: useful for self-signed Vertiv HTTPS endpoints
- `device.type`: recommended explicit device type, one of `ac`, `thd`, `ups`
- `device.equip_id`: CGI request `_equipId` value used for that device
- `exporter.debug_response`: when `true`, parse failures include the full CGI response body in logs

### Device Type Notes

- `ac` devices usually use positive `equip_id` values such as `23`, `24`
- `thd` devices may use request IDs like `-98`
- `ups` devices use their own request `equip_id`, for example `491`

If `type` is omitted, the exporter still has fallback detection heuristics, but explicit `type` is preferred.

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

The exporter does **not** require `docs/vertiv_ac_metrics_list.md` at runtime for normal operation.

AC metric metadata is embedded in [default_metrics.md](/Users/airsmon/Documents/marisme/01_Projects/vertiv_exporter/internal/collector/default_metrics.md:1), so the binary and Docker image can run without `docs/vertiv_ac_metrics_list.md`.

`docs/vertiv_ac_metrics_list.md` is only needed if you want it as a human-maintained reference, or if you explicitly point `exporter.metrics_file` at a custom override file.

The same applies conceptually to:

- `docs/vertiv_thd_metrics_list.md`
- `docs/vertiv_ups_metrics_list.md`

Those files are documentation sources, not required runtime inputs.

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
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache GOPROXY=https://proxy.golang.org,direct go build ./...
```

Start the exporter:

```bash
mkdir -p .gocache .gomodcache
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache GOPROXY=https://proxy.golang.org,direct go run ./cmd/vertiv_exporter -config.file config.yaml
```

Then open:

- `http://127.0.0.1:9101/`
- `http://127.0.0.1:9101/metrics`

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
  -t your-registry/vertiv-exporter:1.0.0 \
  --push .
```

Run it with a mounted config file:

```bash
docker run --rm -p 9101:9101 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  vertiv-exporter:1.0.0
```

The image expects the config file at `/app/config.yaml`.

## Supported Metric Groups

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

- Explicit `device.type` is now supported and recommended. This avoids misclassification when device names vary.
- THD metrics are label-merged instead of exploded into many independent metric names.
- UPS metrics are field-id based and label-driven, which is more stable than parsing English field names.
- Parse errors include a response preview to make field troubleshooting faster.

## Known Behavior

- Prometheus stores numeric values, not display formatting strings.
- For THD metrics, values are rounded to 2 decimal places in code, but `/metrics` may still render `40.4` instead of `40.40`. That is normal Prometheus float formatting behavior.
