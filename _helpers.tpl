{{- /*
  Root helper to build backend FQDN: <svc>.<namespace>.svc.cluster.local:<port>
  Uses values.backend.* from umbrella values (merged subchart values).
*/ -}}
{{- define "app.backendFQDN" -}}
{{- $svcName := default (printf "%s" .Release.Name) .Values.backend.service.name -}}
{{- $svcNS := default .Release.Namespace .Values.backend.namespace -}}
{{- $svcPort := int (default 8080 .Values.backend.service.port) -}}
{{- printf "%s.%s.svc.cluster.local:%d" $svcName $svcNS $svcPort -}}
{{- end -}}