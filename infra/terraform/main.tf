# Cluster ADD-ONS as code (reproducible with a single `terraform apply`).
# The cluster itself is created out-of-band by kind (see Makefile) so the base
# stays zero-cost; the same modules later target EKS in the extension phase.

resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = var.monitoring_namespace
  }
}

resource "helm_release" "kube_prometheus_stack" {
  name       = "kube-prom-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = var.kube_prometheus_stack_version
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  # Keep the demo light on a laptop kind cluster.
  set {
    name  = "grafana.adminPassword"
    value = "admin" # demo only — do not do this in a real environment.
  }
  set {
    name  = "prometheus.prometheusSpec.retention"
    value = "2h"
  }
}

# TODO (Week 4): add a helm_release for Argo CD here too, so the whole control
# plane (GitOps + observability) comes up from one `terraform apply`.
