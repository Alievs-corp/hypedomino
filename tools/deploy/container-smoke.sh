#!/usr/bin/env bash
set -euo pipefail

image="${1:?usage: container-smoke.sh <image>}"
port="${SMOKE_PORT:-18080}"

user=$(docker image inspect -f '{{.Config.User}}' "$image")
[ "$user" = "65532:65532" ] || { echo "image runs as '$user', expected 65532:65532"; exit 1; }

set +e
docker run --rm -e HD_LOG_LEVEL=loud "$image" >/dev/null 2>&1
bad_env_code=$?
set -e
[ "$bad_env_code" = 2 ] || { echo "a bad environment value exited with $bad_env_code, expected 2"; exit 1; }

id=$(docker run -d -p "127.0.0.1:$port:8080" "$image")
trap 'docker rm -f "$id" >/dev/null 2>&1 || true' EXIT

for _ in $(seq 1 50); do
  curl -fsS "http://127.0.0.1:$port/healthz" >/dev/null 2>&1 && break
  sleep 0.2
done
curl -fsS "http://127.0.0.1:$port/healthz"
curl -fsS "http://127.0.0.1:$port/readyz"

started=$(date +%s)
docker stop -t 30 "$id" >/dev/null
elapsed=$(( $(date +%s) - started ))
code=$(docker inspect -f '{{.State.ExitCode}}' "$id")
logs=$(docker logs "$id" 2>&1)
echo "$logs"
[ "$code" = 0 ] || { echo "exit code $code after SIGTERM, expected 0"; exit 1; }
echo "$logs" | grep -q '"msg":"draining"' || { echo "no draining log line"; exit 1; }
echo "$logs" | grep -q '"msg":"stopped"' || { echo "no stopped log line"; exit 1; }
echo "container stopped cleanly in ${elapsed}s"
