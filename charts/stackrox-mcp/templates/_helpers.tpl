{{/*
Create a default fully qualified app name.
*/}}
{{- define "stackrox-mcp.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "stackrox-mcp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "stackrox-mcp.labels" -}}
helm.sh/chart: {{ include "stackrox-mcp.chart" . }}
{{ include "stackrox-mcp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "stackrox-mcp.selectorLabels" -}}
app.kubernetes.io/name: stackrox-mcp
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Image reference
*/}}
{{- define "stackrox-mcp.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository $tag }}
{{- end }}

{{/*
Secret name for API token
*/}}
{{- define "stackrox-mcp.secretName" -}}
{{- if .Values.config.central.existingSecret.name }}
{{- .Values.config.central.existingSecret.name }}
{{- else }}
{{- include "stackrox-mcp.fullname" . }}-api-token
{{- end }}
{{- end }}

{{/*
ConfigMap name
*/}}
{{- define "stackrox-mcp.configMapName" -}}
{{- include "stackrox-mcp.fullname" . }}-config
{{- end }}

{{/*
Detect if running on OpenShift
*/}}
{{- define "stackrox-mcp.isOpenshift" -}}
{{- .Capabilities.APIVersions.Has "route.openshift.io/v1" -}}
{{- end }}
