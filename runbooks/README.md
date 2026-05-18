# Runbooks

Operational runbooks for deploy-pilot. Each alert in
`observability/prometheus/alert-rules.yaml` links to one of these via its
`runbook` annotation.

| Alert | Runbook |
|---|---|
| `DeployPilotHighErrorRate` | [high-error-rate.md](high-error-rate.md) |
| `DeployPilotHighLatencyP99` | [cpu-spike.md](cpu-spike.md) |

A runbook is judged by whether a *different on-call engineer* could resolve the
incident using only the document. Keep steps copy-pasteable.
