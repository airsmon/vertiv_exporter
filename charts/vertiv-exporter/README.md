# Vertiv Exporter Helm Chart

This chart deploys `vertiv_exporter` with secure pod defaults and optional
Prometheus Operator resources.

## Prerequisites

- Kubernetes with Helm
- Network access from the exporter pod to each Vertiv web interface
- An existing Secret containing credentials or a complete `config.yaml`
- Prometheus Operator CRDs when `serviceMonitor.enabled` or
  `prometheusRule.enabled` is enabled

The chart does not create namespaces, credentials, Prometheus Operator, or its
CRDs.

## Install with shared credentials

Create one credential pair used by every configured target:

```bash
kubectl -n monitoring create secret generic vertiv-exporter-credentials \
  --from-literal=username='<VERTIV_USERNAME>' \
  --from-literal=password='<VERTIV_PASSWORD>'
```

Create a values file without credentials:

```yaml
config:
  targets:
    - name: dc-rack-01
      host: https://vertiv.example.invalid
      tlsSkipVerify: false
      devices:
        - name: AC_1
          type: ac
          equipId: 23
        - name: ENV_THD
          type: thd
          equipId: -98
        - name: UPS_1
          type: ups
          equipId: 26

credentials:
  existingSecret:
    name: vertiv-exporter-credentials
```

Install the chart:

```bash
helm upgrade --install vertiv-exporter ./charts/vertiv-exporter \
  --namespace monitoring \
  --create-namespace \
  --values vertiv-values.yaml
```

Do not place usernames or passwords in Helm values. Helm stores release values
in the cluster.

## Install with a complete configuration Secret

Use this mode when targets require different credentials. Create the Secret
from a local configuration file:

```bash
kubectl -n monitoring create secret generic vertiv-exporter-config \
  --from-file=config.yaml=./config.yaml
```

Reference it from values:

```yaml
config:
  existingSecret:
    name: vertiv-exporter-config
    key: config.yaml
```

When an external configuration Secret is used, its `exporter.listen_address`
and `exporter.metrics_path` must match `exporter.port` and
`exporter.metricsPath` in the Helm values. Restart the Deployment after changing
the external Secret because Helm cannot calculate a checksum for resources it
does not manage. The same restart requirement applies after changing the shared
credential Secret.

## Prometheus Operator

Enable ServiceMonitor discovery and built-in alerts:

```yaml
serviceMonitor:
  enabled: true
  labels:
    prometheus: kube-prometheus

prometheusRule:
  enabled: true
  labels:
    prometheus: kube-prometheus
  ruleLabels:
    team: infrastructure
```

`serviceMonitor.labels` and `prometheusRule.labels` must match the selectors
used by the installed Prometheus resource.

The built-in rules are:

| Alert | Default condition | For | Severity |
| --- | --- | --- | --- |
| `VertivExporterTargetDown` | `vertiv_exporter_up == 0` | `5m` | critical |
| `VertivTHDCommunicationFault` | `vertiv_thd_comm_status != 0` | `5m` | critical |
| `VertivTHDHighTemperature` | `vertiv_thd_high_temp_alarm_rack_count > 0` | `2m` | critical |
| `VertivTHDDoorOpen` | `vertiv_thd_door_status == 1` | `5m` | warning |
| `VertivUPSUtilityPowerLost` | `vertiv_ups_status_power_supply != 1` | `1m` | critical |
| `VertivUPSHighLoad` | `vertiv_ups_output_load_percent > 90` | `10m` | warning |
| `VertivUPSLowBatteryCapacity` | `vertiv_ups_battery_capacity_percent < 20` | `5m` | critical |
| `VertivACActiveAlarm` | `vertiv_ac_system_alarm_active_count > 0` | `5m` | warning |

Every built-in rule can be disabled or adjusted under
`prometheusRule.rules`. Complete custom groups can be appended with
`prometheusRule.additionalGroups`.

## Configuration notes

- `replicaCount` defaults to one because each replica maintains independent
  device login sessions and emits duplicate metrics.
- The Deployment strategy defaults to `Recreate` for the same reason.
- Liveness, readiness, and startup probes use `/`; they do not initiate device
  collection through `/metrics`.
- The generated ConfigMap contains target addresses and device IDs but never
  credentials.
- `checksum/config` rolls pods when the generated ConfigMap changes.
- The default pod runs as UID/GID `65532`, uses a read-only root filesystem,
  disables privilege escalation, drops all Linux capabilities, and does not
  mount a ServiceAccount token.

## Validate

```bash
helm lint charts/vertiv-exporter
helm lint charts/vertiv-exporter \
  --values charts/vertiv-exporter/ci/test-values.yaml
helm template vertiv-exporter charts/vertiv-exporter \
  --namespace monitoring \
  --values charts/vertiv-exporter/ci/test-values.yaml
helm package charts/vertiv-exporter
```

Defaults intentionally contain no target or credentials. Supply one of the
configuration modes above before deploying.
