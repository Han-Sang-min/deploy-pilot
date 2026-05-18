# deploy-pilot — GitOps Reference Platform

> A commit becomes a **verified, observed, recoverable** Kubernetes deployment —
> with zero manual `kubectl`. One repo that demonstrates the full junior-DevOps
> loop: IaC → CI/CD → GitOps → observability → documented failure response.

![ci](https://img.shields.io/badge/CI-pending-lightgrey)
`Go` · `Docker` · `GitHub Actions` · `Kubernetes` · `Kustomize` · `Argo CD` ·
`Terraform` · `Prometheus` · `Grafana`

> **Status:** Week-1 scaffold. The roadmap below maps every section to a
> concrete artifact in this repo. Boxes are checked as evidence (screenshots /
> badges / links) lands.

---

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Quick Start](#quick-start)
4. [CI/CD Pipeline](#cicd-pipeline)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Terraform Infrastructure](#terraform-infrastructure)
7. [Observability](#observability)
8. [Failure Scenarios](#failure-scenarios)
9. [Troubleshooting / Runbooks](#troubleshooting--runbooks)
10. [Roadmap](#roadmap)
11. [Lessons Learned](#lessons-learned)
12. [Reused from go-agent-core](#reused-from-go-agent-core)

---

## Overview
A deliberately small Go HTTP service is the *payload*; the project is about the
**pipeline around it**. What a reviewer should be able to verify in 2 minutes:
push → CI gates (test + Trivy) → image to GHCR → Argo CD syncs the cluster →
Grafana shows RED metrics → an injected failure fires an alert → a runbook
restores service → a postmortem closes the loop.

> _Evidence to add here: a 20s demo GIF of the full loop._

## Architecture
See [docs/architecture.md](docs/architecture.md) for the full diagram and
component boundaries.

## Quick Start
```bash
# Run the service locally
make run                       # :8080  → /healthz /readyz /metrics /work /boom /spin
curl -s localhost:8080/healthz
curl -s localhost:8080/metrics | grep http_requests_total

# Bring up the platform on a local kind cluster
make kind-up
make kind-load
make argocd-install
make deploy                    # Argo CD reconciles deploy/k8s/overlays/dev
make observability-install     # or: make tf-apply  (same stack via Terraform)
```

## CI/CD Pipeline
[.github/workflows/ci.yaml](.github/workflows/ci.yaml) — two stages:
- **Lint & Test:** `gofmt` check, `go vet`, `go test -race`.
- **Build, Scan & Push:** multi-stage build → **Trivy gate on HIGH/CRITICAL** →
  push to GHCR on `main`. GitOps tag-bump step is scaffolded (Week 2).

> _Evidence to add: green CI badge (swap the placeholder above), Trivy output._

## Kubernetes Deployment
[deploy/k8s](deploy/k8s) — Kustomize `base` + `dev`/`prod` overlays. The base is
intentionally production-grade (this is the cheapest high-signal artifact):
liveness/readiness probes, resource requests/limits, non-root +
`readOnlyRootFilesystem` + dropped capabilities + seccomp, HPA, PodDisruptionBudget,
topology spread, `maxUnavailable: 0` rolling updates, SIGTERM-aware graceful drain.

## Terraform Infrastructure
[infra/terraform](infra/terraform/README.md) — cluster **add-ons as code**
(`kube-prometheus-stack`), reproducible via `make tf-apply`. Cluster creation
stays in kind to keep the base zero-cost; the same modules retarget EKS in the
extension phase.

## Observability
[observability/](observability) — `ServiceMonitor` scrape config, a versioned
[Grafana RED dashboard](observability/grafana/dashboards/service-red.json),
`PrometheusRule` SLO alerts, and an `AlertmanagerConfig`. Metrics are standard
Prometheus exposition via `client_golang`.

> _Evidence to add: Grafana RED screenshot under normal + degraded load._

## Failure Scenarios
[tools/faultinject/faultinject.sh](tools/faultinject/faultinject.sh) drives two
reproducible scenarios through the deployed service:

| Scenario | Command | Alert | Runbook |
|---|---|---|---|
| High error rate | `make fault-errors` | `DeployPilotHighErrorRate` | [high-error-rate.md](runbooks/high-error-rate.md) |
| CPU spike / latency | `make fault-cpu` | `DeployPilotHighLatencyP99` | [cpu-spike.md](runbooks/cpu-spike.md) |

> _Evidence to add: timeline table — inject → alert → mitigate → recover._

## Troubleshooting / Runbooks
[runbooks/](runbooks/README.md) — copy-pasteable, alert-linked.
[postmortems/](postmortems/TEMPLATE.md) — blameless template; one real
postmortem lands in Week 4.

## Roadmap
**1-month MVP (kind, zero cost):**
- [x] **W1 — Build + CI:** Go service, multi-stage Docker, GitHub Actions (test + Trivy)
- [ ] **W2 — Cluster + GitOps:** prod-grade manifests, Argo CD auto-sync, CI tag-bump
- [ ] **W3 — Observability:** Prometheus/Grafana, RED dashboard, alert rules
- [ ] **W4 — IaC + failure + docs:** Terraform add-ons, 1 injected incident, runbook + postmortem, README evidence

**2–3 month extension:** EKS via Terraform · staging→prod promotion · Trivy as
hard merge gate · go-agent-core /proc collector as a Prometheus-exporter DaemonSet.

## Lessons Learned
> _Fill as you go: design trade-offs, what broke, what you'd do differently._
> First entry to capture: the Argo CD `selfHeal` vs `kubectl rollout undo`
> conflict (see [runbooks/high-error-rate.md](runbooks/high-error-rate.md)).

## Reused from go-agent-core
This project is the new headline; `go-agent-core` is demoted to a supporting
component. Reused: the **failure-injection concept** (here as `faultinject.sh`).
Planned reuse: its `/proc`+cgroup collector becomes a Prometheus-exporter
DaemonSet in the extension phase — turning "hand-rolled /proc parsing" from a
weakness into a "knows the standard stack *and* the internals" strength.
