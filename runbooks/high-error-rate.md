# Runbook: High Error Rate (5xx > 5%)

**Alert:** `DeployPilotHighErrorRate` · **Severity:** critical

## 1. Symptom
5xx ratio over 5% for 5m. Users see failed requests.

## 2. Confirm (60s)
- Grafana → *deploy-pilot — RED* → "Error ratio" panel trending up.
- ```bash
  kubectl -n deploy-pilot get pods -l app.kubernetes.io/name=deploy-pilot
  kubectl -n deploy-pilot logs -l app.kubernetes.io/name=deploy-pilot --tail=100 | grep -i error
  ```

## 3. Triage
| Check | Command | If true → |
|---|---|---|
| Bad recent rollout? | `kubectl -n deploy-pilot rollout history deploy/deploy-pilot` | go to Mitigation A |
| Crash-looping pods? | `kubectl -n deploy-pilot get pods` (look for `CrashLoopBackOff`) | go to Mitigation B |
| Synthetic load (`/boom`)? | check `tools/faultinject` is not running | stop the injector |

## 4. Mitigation
**A — roll back the last deploy (fastest):**
```bash
kubectl -n deploy-pilot rollout undo deploy/deploy-pilot
kubectl -n deploy-pilot rollout status deploy/deploy-pilot --timeout=120s
```
> Note: with Argo CD `selfHeal: true`, also pin the previous image tag in
> `deploy/k8s/overlays/dev/kustomization.yaml` and commit, or Argo re-syncs the
> bad version. This is the GitOps trade-off — record it in the postmortem.

**B — restart unhealthy pods:**
```bash
kubectl -n deploy-pilot rollout restart deploy/deploy-pilot
```

## 5. Verify
Error ratio panel returns below 5% and stays for 5m; alert resolves.

## 6. Aftermath
Open a postmortem from `postmortems/TEMPLATE.md` within 24h.
