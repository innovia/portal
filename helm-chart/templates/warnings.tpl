{{/* this file is for generating warnings about incorrect usage of the chart */}}

{{- if not .Values.mtls.ca_cert  }}
  {{ fail "mtls.ca_cert value should be set to a content of file and should not empty "}}
{{- end }}

{{- if not .Values.mtls.server_key  }}
  {{ fail "mtls.server_key value should be set to a content of file and should not empty "}}
{{- end }}

{{- if not .Values.mtls.server_cert  }}
  {{ fail "mtls.server_cert value should be set to a content of file and should not empty "}}
{{- end }}
