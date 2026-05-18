variable "kubeconfig" {
  description = "Path to the kubeconfig file."
  type        = string
  default     = "~/.kube/config"
}

variable "kube_context" {
  description = "kube context to target (kind creates kind-deploy-pilot)."
  type        = string
  default     = "kind-deploy-pilot"
}

variable "monitoring_namespace" {
  description = "Namespace for the kube-prometheus-stack release."
  type        = string
  default     = "monitoring"
}

variable "kube_prometheus_stack_version" {
  description = "Chart version for kube-prometheus-stack (pin for reproducibility)."
  type        = string
  default     = "65.1.1"
}
