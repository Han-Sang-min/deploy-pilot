# infra/terraform

Cluster **add-ons** as code. The kind cluster is created by `make kind-up`
(zero cost, laptop-friendly); Terraform then installs the observability stack
reproducibly.

```bash
make kind-up
make tf-init
make tf-apply        # installs kube-prometheus-stack into the monitoring ns
terraform output     # prints the Grafana port-forward command
```

**Why cluster creation is not in Terraform (yet):** keeping bootstrap in kind
keeps the base demo free and fast. The Week-6+ extension swaps these same
provider/helm modules onto an EKS cluster created by a `terraform-aws-eks`
module — the add-on code does not change, which is the point.
