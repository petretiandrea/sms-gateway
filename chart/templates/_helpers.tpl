{{/*
Expand the chart name.
*/}}
{{- define "sms-gateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "sms-gateway.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "sms-gateway.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | quote }}
app.kubernetes.io/name: {{ include "sms-gateway.name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- end -}}

{{/*
Selector labels for a component.
*/}}
{{- define "sms-gateway.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sms-gateway.name" .root | quote }}
app.kubernetes.io/instance: {{ .root.Release.Name | quote }}
app.kubernetes.io/component: {{ .component | quote }}
{{- end -}}

{{/*
Generated resource names.
*/}}
{{- define "sms-gateway.configMapName" -}}
{{ include "sms-gateway.fullname" . }}-config
{{- end -}}

{{- define "sms-gateway.commonSecretName" -}}
{{ include "sms-gateway.fullname" . }}-secret
{{- end -}}

{{- define "sms-gateway.secretFilesName" -}}
{{ include "sms-gateway.fullname" . }}-secret-files
{{- end -}}

{{- define "sms-gateway.componentSecretName" -}}
{{ include "sms-gateway.fullname" .root }}-{{ .component }}-secret
{{- end -}}

{{/*
Common envFrom sources inherited by all runtime components.
*/}}
{{- define "sms-gateway.commonEnvFrom" -}}
{{- if .Values.common.config.env }}
- configMapRef:
    name: {{ include "sms-gateway.configMapName" . }}
{{- end }}
{{- range .Values.common.config.existingConfigMaps }}
- configMapRef:
    name: {{ . | quote }}
{{- end }}
{{- if .Values.common.secrets.env }}
- secretRef:
    name: {{ include "sms-gateway.commonSecretName" . }}
{{- end }}
{{- range .Values.common.secrets.existingSecrets }}
- secretRef:
    name: {{ . | quote }}
{{- end }}
{{- end -}}

{{/*
Component-specific Secret envFrom source.
*/}}
{{- define "sms-gateway.componentSecretEnvFrom" -}}
{{- if .values.secrets }}
- secretRef:
    name: {{ include "sms-gateway.componentSecretName" . }}
{{- end }}
{{- end -}}

{{/*
Common env vars mapped from specific keys in existing Secrets.
*/}}
{{- define "sms-gateway.commonSecretKeyEnvVars" -}}
{{- range $name, $ref := .Values.common.secrets.envFromKeys }}
- name: {{ $name }}
  valueFrom:
    secretKeyRef:
      name: {{ $ref.secretName | quote }}
      key: {{ $ref.key | quote }}
{{- end }}
{{- end -}}

{{/*
Map-style component env vars.
*/}}
{{- define "sms-gateway.componentEnvVars" -}}
{{- range $name, $value := . }}
- name: {{ $name }}
  value: {{ $value | quote }}
{{- end }}
{{- end -}}

{{/*
Common volume mounts inherited by all runtime components.
*/}}
{{- define "sms-gateway.commonVolumeMounts" -}}
{{- if .Values.common.secrets.files }}
- name: common-secret-files
  mountPath: /etc/sms-gateway/secrets
  readOnly: true
{{- end }}
{{- range $index, $secret := .Values.common.secrets.existingSecretFiles }}
- name: existing-secret-files-{{ $index }}
  mountPath: {{ $secret.mountPath | quote }}
  readOnly: {{ if hasKey $secret "readOnly" }}{{ $secret.readOnly }}{{ else }}true{{ end }}
{{- end }}
{{- end -}}

{{/*
Common volumes inherited by all runtime components.
*/}}
{{- define "sms-gateway.commonVolumes" -}}
{{- if .Values.common.secrets.files }}
- name: common-secret-files
  secret:
    secretName: {{ include "sms-gateway.secretFilesName" . }}
{{- end }}
{{- range $index, $secret := .Values.common.secrets.existingSecretFiles }}
- name: existing-secret-files-{{ $index }}
  secret:
    secretName: {{ $secret.name | quote }}
{{- end }}
{{- end -}}
