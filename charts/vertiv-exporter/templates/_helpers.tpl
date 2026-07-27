{{/*
Expand the chart name.
*/}}
{{- define "vertiv-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a fully qualified application name.
*/}}
{{- define "vertiv-exporter.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create the chart name and version used by the chart label.
*/}}
{{- define "vertiv-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "vertiv-exporter.labels" -}}
{{- $labels := mergeOverwrite (dict) .Values.commonLabels (include "vertiv-exporter.selectorLabels" . | fromYaml) (dict
  "helm.sh/chart" (include "vertiv-exporter.chart" .)
  "app.kubernetes.io/managed-by" .Release.Service
  "app.kubernetes.io/component" "exporter"
) }}
{{- if .Chart.AppVersion }}
{{- $_ := set $labels "app.kubernetes.io/version" .Chart.AppVersion }}
{{- end }}
{{- toYaml $labels }}
{{- end }}

{{/*
Stable selector labels.
*/}}
{{- define "vertiv-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vertiv-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name.
*/}}
{{- define "vertiv-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "vertiv-exporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Container image reference.
*/}}
{{- define "vertiv-exporter.image" -}}
{{- if .Values.image.digest }}
{{- printf "%s@%s" .Values.image.repository .Values.image.digest }}
{{- else }}
{{- printf "%s:%s" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) }}
{{- end }}
{{- end }}

{{/*
Configuration resource name.
*/}}
{{- define "vertiv-exporter.configName" -}}
{{- default (include "vertiv-exporter.fullname" .) .Values.config.existingSecret.name }}
{{- end }}
