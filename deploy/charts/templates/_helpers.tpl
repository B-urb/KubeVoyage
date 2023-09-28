{{/* Generate the best default app name */}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
