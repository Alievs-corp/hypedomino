#!/usr/bin/env bash
set -euo pipefail

cluster="${1:?usage: kind-rollout.sh <kind cluster> [image]}"
image="${2:-hypedomino-server:ci}"
namespace=hd-ci
rollback_budget_seconds=300
work=$(mktemp -d)
watcher=""
trap '[ -n "$watcher" ] && kill "$watcher" 2>/dev/null; true' EXIT

kind load docker-image "$image" --name "$cluster"
kubectl apply -k deploy/k8s/overlays/ci
kubectl -n "$namespace" rollout status deploy/server --timeout=180s
replicas=$(kubectl -n "$namespace" get deploy/server -o jsonpath='{.spec.replicas}')

old_pods=$(kubectl -n "$namespace" get pods -l app.kubernetes.io/name=server -o jsonpath='{.items[*].metadata.name}')
followers=()
for pod in $old_pods; do
  kubectl -n "$namespace" logs -f "$pod" >"$work/$pod.log" 2>&1 &
  followers+=($!)
done

watch_available() {
  while true; do
    kubectl -n "$namespace" get deploy/server -o jsonpath='{.status.availableReplicas}{"\n"}' >>"$work/available.txt" 2>/dev/null || true
    sleep 1
  done
}
watch_available &
watcher=$!

kubectl -n "$namespace" rollout restart deploy/server
kubectl -n "$namespace" rollout status deploy/server --timeout=300s
for pid in "${followers[@]}"; do
  timeout 120 tail --pid="$pid" -f /dev/null || { echo "an old pod did not stop within 120s"; exit 1; }
done
kill "$watcher"
watcher=""

min_available=$(grep -E '^[0-9]+$' "$work/available.txt" | sort -n | head -1)
echo "available replicas during the rollout: min ${min_available:-none} of $replicas"
[ "${min_available:-0}" -ge "$replicas" ] || { echo "capacity dropped below $replicas during the rollout"; exit 1; }

for pod in $old_pods; do
  echo "--- $pod"
  cat "$work/$pod.log"
  grep -q '"msg":"draining"' "$work/$pod.log" || { echo "$pod did not drain"; exit 1; }
  grep -q '"msg":"stopped"' "$work/$pod.log" || { echo "$pod did not stop cleanly"; exit 1; }
done

started=$(date +%s)
kubectl -n "$namespace" rollout undo deploy/server
kubectl -n "$namespace" rollout status deploy/server --timeout="${rollback_budget_seconds}s"
elapsed=$(( $(date +%s) - started ))
echo "rollback serving on the previous revision after ${elapsed}s (budget ${rollback_budget_seconds}s)"
[ "$elapsed" -lt "$rollback_budget_seconds" ]
