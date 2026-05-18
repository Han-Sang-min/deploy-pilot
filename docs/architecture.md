# Architecture

deploy-pilot proves a single end-to-end DevOps narrative: **a commit becomes a
verified, observed, recoverable deployment with no manual `kubectl`.**

```
 Developer
    │  git push / PR
    ▼
┌──────────────────────── GitHub Actions (CI) ───────────────────────┐
│  gofmt · go vet · go test -race                                    │
│  docker build (multi-stage, distroless nonroot)                    │
│  Trivy scan  ── HIGH/CRITICAL ⇒ fail (merge gate)                  │
│  push image → GHCR                                                  │
│  [Week 2] bump image tag in overlays/dev/kustomization.yaml ───────┐│
└────────────────────────────────────────────────────────────────────┘│
                                                                       │ git commit
                                                                       ▼
                                                            ┌────────────────────┐
                                                            │  Git (desired      │
                                                            │  state / manifests)│
                                                            └─────────┬──────────┘
                                                                      │ watch
                                                                      ▼
                          ┌──────────────── Argo CD ──────────────────┐
                          │  sync (prune + selfHeal)                   │
                          └─────────────────────┬──────────────────────┘
                                                ▼
   ┌──────────────────── kind cluster (laptop) / EKS (extension) ─────────────────┐
   │  Deployment (probes, limits, securityContext, HPA, PDB, topologySpread)      │
   │  Service ──/metrics──▶ ServiceMonitor ──▶ Prometheus ──▶ Grafana (RED)       │
   │                                              │                               │
   │                                       Alertmanager ──▶ runbook ──▶ recovery  │
   └──────────────────────────────────────────────────────────────────────────────┘
            ▲
            │ tools/faultinject.sh  (reproducible cpu-spike / error scenarios)
```

## Component boundaries
- **App** (`cmd/`, `internal/`) — small instrumented HTTP service; the *thing
  being deployed*, deliberately minimal.
- **deploy/k8s** — Kustomize base + dev/prod overlays (the desired state).
- **gitops/argocd** — what continuously reconciles the cluster to Git.
- **infra/terraform** — cluster add-ons as code (observability stack).
- **observability/** — scrape rules, alert rules, dashboards (versioned).
- **runbooks/ + postmortems/** — operability as a first-class artifact.

## Reused from go-agent-core
The failure-injection idea comes from go-agent-core's simulator. The /proc &
cgroup collector from that project is the planned **Prometheus-exporter
DaemonSet** in the extension phase — "knows standard tooling *and* internals".
