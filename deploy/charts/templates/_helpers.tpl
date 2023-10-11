{{/* Generate the best default app name */}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}

{{- define "kubevoyage.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubevoyage.labels" -}}
app.kubernetes.io/name: {{ include "kubevoyage.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}