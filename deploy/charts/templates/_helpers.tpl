{{/* Generate the best default app name */}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}

{{- define "kubevoyage.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
