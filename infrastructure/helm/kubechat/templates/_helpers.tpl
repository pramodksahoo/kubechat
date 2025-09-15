{{/*
Expand the name of the chart.
*/}}
{{- define "kubechat.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kubechat.fullname" -}}
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
{{- define "kubechat.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubechat.labels" -}}
helm.sh/chart: {{ include "kubechat.chart" . }}
{{ include "kubechat.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubechat.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubechat.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Web app labels
*/}}
{{- define "kubechat.web.labels" -}}
{{ include "kubechat.labels" . }}
app.kubernetes.io/component: web
{{- end }}

{{/*
Web app selector labels
*/}}
{{- define "kubechat.web.selectorLabels" -}}
{{ include "kubechat.selectorLabels" . }}
app.kubernetes.io/component: web
{{- end }}

{{/*
API app labels
*/}}
{{- define "kubechat.api.labels" -}}
{{ include "kubechat.labels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
API app selector labels
*/}}
{{- define "kubechat.api.selectorLabels" -}}
{{ include "kubechat.selectorLabels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
Ollama app labels
*/}}
{{- define "kubechat.ollama.labels" -}}
{{ include "kubechat.labels" . }}
app.kubernetes.io/component: ollama
{{- end }}

{{/*
Ollama selector labels
*/}}
{{- define "kubechat.ollama.selectorLabels" -}}
{{ include "kubechat.selectorLabels" . }}
app.kubernetes.io/component: ollama
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kubechat.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kubechat.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the namespace name
*/}}
{{- define "kubechat.namespace" -}}
{{- default .Release.Namespace .Values.namespace.name }}
{{- end }}

{{/*
Web image name
*/}}
{{- define "kubechat.web.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.web.image.registry -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .Values.web.image.repository (.Values.web.image.tag | default .Chart.AppVersion) -}}
{{- else }}
{{- printf "%s:%s" .Values.web.image.repository (.Values.web.image.tag | default .Chart.AppVersion) -}}
{{- end }}
{{- end }}

{{/*
API image name
*/}}
{{- define "kubechat.api.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.api.image.registry -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .Values.api.image.repository (.Values.api.image.tag | default .Chart.AppVersion) -}}
{{- else }}
{{- printf "%s:%s" .Values.api.image.repository (.Values.api.image.tag | default .Chart.AppVersion) -}}
{{- end }}
{{- end }}

{{/*
PostgreSQL connection string
*/}}
{{- define "kubechat.postgresql.connectionString" -}}
{{- if .Values.postgresql.enabled -}}
postgres://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ include "kubechat.fullname" . }}-postgresql:5432/{{ .Values.postgresql.auth.database }}?sslmode=disable
{{- else -}}
{{- .Values.externalDatabase.connectionString -}}
{{- end -}}
{{- end }}

{{/*
Redis connection string
*/}}
{{- define "kubechat.redis.connectionString" -}}
{{- if .Values.redis.enabled -}}
{{- if .Values.redis.auth.enabled -}}
redis://:{{ .Values.redis.auth.password }}@{{ include "kubechat.fullname" . }}-redis-master:6379
{{- else -}}
redis://{{ include "kubechat.fullname" . }}-redis-master:6379
{{- end -}}
{{- else -}}
{{- .Values.externalRedis.connectionString -}}
{{- end -}}
{{- end }}