output "monitoring_namespace" {
  description = "Namespace where the observability stack is installed."
  value       = kubernetes_namespace.monitoring.metadata[0].name
}

output "grafana_port_forward" {
  description = "Command to reach Grafana locally."
  value       = "kubectl -n ${var.monitoring_namespace} port-forward svc/kube-prom-stack-grafana 3000:80"
}
