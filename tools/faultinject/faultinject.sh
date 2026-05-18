#!/usr/bin/env bash
# Reproducible failure injection for the deploy-pilot demo.
# Reuses the failure-scenario idea from go-agent-core's simulator, but drives it
# through the deployed service so alerts + runbooks can be demonstrated for real.
#
# Usage:
#   ./faultinject.sh errors [BASE_URL] [SECONDS]   # drives DeployPilotHighErrorRate
#   ./faultinject.sh cpu    [BASE_URL] [SECONDS]   # drives DeployPilotHighLatencyP99
set -euo pipefail

MODE="${1:-errors}"
BASE_URL="${2:-http://localhost:8080}"
SECONDS_TO_RUN="${3:-360}"
END=$(( $(date +%s) + SECONDS_TO_RUN ))

echo "[faultinject] mode=$MODE target=$BASE_URL duration=${SECONDS_TO_RUN}s"

case "$MODE" in
  errors)
    while [ "$(date +%s)" -lt "$END" ]; do
      curl -fsS -o /dev/null "$BASE_URL/boom" || true
      curl -fsS -o /dev/null "$BASE_URL/work?ms=20" || true
      sleep 0.2
    done
    ;;
  cpu)
    while [ "$(date +%s)" -lt "$END" ]; do
      for _ in 1 2 3 4 5 6 7 8; do
        curl -fsS -o /dev/null "$BASE_URL/spin?ms=400" &
      done
      wait
    done
    ;;
  *)
    echo "unknown mode: $MODE (use 'errors' or 'cpu')" >&2
    exit 2
    ;;
esac

echo "[faultinject] done — check Grafana and Alertmanager, then follow the runbook."
