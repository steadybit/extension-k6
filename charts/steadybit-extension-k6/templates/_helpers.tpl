{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "k6.secret.name" -}}
{{- default "steadybit-extension-k6" .Values.k6.existingSecret -}}
{{- end -}}
