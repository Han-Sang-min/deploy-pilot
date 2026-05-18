# Runbook: CPU Spike / High p99 Latency

**Alert:** `DeployPilotHighLatencyP99` · **Severity:** warning

## 1. Symptom
p99 latency over 1s for 10m. Requests slow; HPA may be scaling.

## 2. Confirm (60s)
- Grafana → *deploy-pilot — RED* → "Latency p50/p99" panel.
- ```bash
  kubectl -n deploy-pilot top pods
  kubectl -n deploy-pilot get hpa deploy-pilot
  ```

## 3. Triage
| Check | Command | If true → |
|---|---|---|
| Pods CPU-throttled at limit? | `kubectl -n deploy-pilot top pods` near `250m` | Mitigation A |
| HPA maxed out? | `kubectl -n deploy-pilot get hpa` shows `REPLICAS=6` | Mitigation B |
| Synthetic load (`/spin`)? | check `tools/faultinject` is not running | stop the injector |

## 4. Mitigation
**A — give headroom (temporary):**
```bash
kubectl -n deploy-pilot set resources deploy/deploy-pilot \
  --limits=cpu=500m --requests=cpu=100m
```
**B — raise the autoscaler ceiling (temporary):**
```bash
kubectl -n deploy-pilot patch hpa deploy-pilot \
  --type=merge -p '{"spec":{"maxReplicas":10}}'
```
> Both are stopgaps. The durable fix (right-sized limits / code hotspot) goes
> into the postmortem's prevention section, then back into Git.

## 5. Verify
p99 panel back under 1s for 10m; HPA stabilises; alert resolves.

## 6. Aftermath
Postmortem within 24h; fold the durable fix into `deploy/k8s/base`.
