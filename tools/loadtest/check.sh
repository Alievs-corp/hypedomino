#!/usr/bin/env bash
set -euo pipefail

run_dir="${1:?usage: check.sh <run dir with vmstat.log copied in> [ramp minutes] [hold minutes]}"
ramp="${2:-5}"
hold="${3:-30}"
started=$(cat "$run_dir/started")
failed=0

heap_at() {
  local want=$(( started + ($1) * 60 ))
  local file
  file=$(ls "$run_dir"/metrics/*.prom | awk -F/ -v want="$want" '{ n = $NF; sub(/\.prom$/, "", n); d = n - want; if (d < 0) d = -d; if (best == "" || d < bestd) { best = $0; bestd = d } } END { print best }')
  awk '$1 == "go_heap_live_bytes" { print $2 }' "$file"
}

early=$(heap_at "$ramp + 10")
late=$(heap_at "$ramp + $hold")
drift=$(awk -v a="$early" -v b="$late" 'BEGIN { printf "%.2f", (b - a) * 100 / a }')
echo "heap minute 10: $early bytes, minute $hold: $late bytes, drift $drift%"
awk -v d="$drift" 'BEGIN { exit !(d > 5 || d < -5) }' && { echo "FAIL heap drift over 5%"; failed=1; }

cpu=$(awk -v skip="$(( ramp * 6 + 2 ))" -v take="$(( hold * 6 ))" 'NR > skip && NR <= skip + take { busy += 100 - $15; n++ } END { if (n) printf "%.1f", busy / n }' "$run_dir/vmstat.log")
echo "average CPU during the hold: $cpu%"
awk -v c="$cpu" 'BEGIN { exit !(c == "" || c >= 80) }' && { echo "FAIL average CPU not under 80%"; failed=1; }

awk -F'[{}" ]+' '
  $1 == "move_roundtrip_seconds_bucket" { bound[n] = $3; cum[n] = $NF; n++ }
  $1 == "move_roundtrip_seconds_count" { count = $2 }
  END {
    for (i = 0; i < n; i++) if (cum[i] >= count * 0.99) { print "server move_roundtrip_seconds p99 <= " bound[i] " s over " count " moves"; exit }
  }' "$run_dir/final.prom"

exit "$failed"
