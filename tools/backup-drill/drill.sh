#!/usr/bin/env bash
set -euo pipefail
export MSYS_NO_PATHCONV=1

here=$(cd "$(dirname "$0")" && pwd)
run=$(date +%s)-$$
image="hd-backup-drill:local"
network="hd-drill-$run"
objectstore="hd-drill-objectstore-$run"
source_db="hd-drill-source-$run"
restored_db="hd-drill-restored-$run"
source_volume="hd-drill-source-data-$run"
restored_volume="hd-drill-restored-data-$run"
bucket="hd-backups"
access_key="drill$(openssl rand -hex 6)"
secret_key="$(openssl rand -hex 16)"
encryption_key="$(openssl rand -hex 32)"

cleanup() {
  docker rm -f "$source_db" "$restored_db" "$objectstore" >/dev/null 2>&1 || true
  docker volume rm "$source_volume" "$restored_volume" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
}
trap cleanup EXIT

walg_env=(
  -e "WALG_S3_PREFIX=s3://$bucket/drill"
  -e "AWS_ENDPOINT=http://$objectstore:9000"
  -e "AWS_S3_FORCE_PATH_STYLE=true"
  -e "AWS_REGION=us-east-1"
  -e "AWS_ACCESS_KEY_ID=$access_key"
  -e "AWS_SECRET_ACCESS_KEY=$secret_key"
  -e "WALG_LIBSODIUM_KEY=$encryption_key"
  -e "WALG_LIBSODIUM_KEY_TRANSFORM=hex"
  -e "WALG_COMPRESSION_METHOD=zstd"
  -e "PGUSER=postgres"
  -e "PGHOST=/var/run/postgresql"
)

sql() {
  docker exec -u postgres "$1" psql -qtAX -v ON_ERROR_STOP=1 -c "$2"
}

wait_ready() {
  for _ in $(seq 1 120); do
    docker exec -u postgres "$1" pg_isready -q -h /var/run/postgresql && return 0
    sleep 0.5
  done
  echo "$1 did not become ready"
  docker logs "$1" 2>&1 | tail -30
  return 1
}

elapsed_since() {
  echo $(( $(date +%s) - $1 ))
}

(cd "$here" && docker build -q -t "$image" .) >/dev/null
docker network create "$network" >/dev/null

docker run -d --name "$objectstore" --network "$network"   -e "AWS_ACCESS_KEY_ID=$access_key" -e "AWS_SECRET_ACCESS_KEY=$secret_key"   chrislusf/seaweedfs:4.53 server -s3 -s3.port=9000 -dir=/data >/dev/null
docker run --rm --network "$network" --entrypoint bash "$image" -c "
  for i in \$(seq 1 60); do
    curl -sS -o /dev/null --aws-sigv4 aws:amz:us-east-1:s3 --user $access_key:$secret_key -X PUT http://$objectstore:9000/$bucket && exit 0
    sleep 1
  done
  exit 1" >/dev/null

docker volume create "$source_volume" >/dev/null
docker run -d --name "$source_db" --network "$network" "${walg_env[@]}" \
  -e POSTGRES_PASSWORD=drill -v "$source_volume:/var/lib/postgresql/data" "$image" \
  postgres -c wal_level=replica -c archive_mode=on -c archive_timeout=60 \
  -c "archive_command=wal-g wal-push %p" >/dev/null
wait_ready "$source_db"

sql "$source_db" "create table ledger (id bigserial primary key, batch text not null, amount int not null, at timestamptz not null default clock_timestamp())"
sql "$source_db" "insert into ledger (batch, amount) select 'before-backup', g from generate_series(1, 1000) g"
docker exec -u postgres "${walg_env[@]}" "$source_db" wal-g backup-push /var/lib/postgresql/data >/dev/null

sql "$source_db" "insert into ledger (batch, amount) select 'after-backup', g from generate_series(1, 500) g"
sql "$source_db" "select pg_switch_wal()" >/dev/null
sleep 2
target_time=$(sql "$source_db" "select to_char(clock_timestamp() at time zone 'UTC', 'YYYY-MM-DD HH24:MI:SS.US') || '+00'")
sleep 2
sql "$source_db" "insert into ledger (batch, amount) select 'after-target', g from generate_series(1, 250) g"
sql "$source_db" "select pg_switch_wal()" >/dev/null

for _ in $(seq 1 60); do
  archived=$(sql "$source_db" "select coalesce(last_archived_wal >= pg_walfile_name(pg_current_wal_lsn() - 1048576), false) from pg_stat_archiver")
  [ "$archived" = "t" ] && break
  sleep 1
done
[ "$archived" = "t" ] || { echo "the last WAL segment was not archived in time"; exit 1; }
failed=$(sql "$source_db" "select failed_count from pg_stat_archiver")
[ "$failed" = "0" ] || { echo "WAL archiving failed $failed times"; exit 1; }
echo "source rows: $(sql "$source_db" "select count(*) from ledger") (recovery target $target_time)"
docker rm -f "$source_db" >/dev/null

docker volume create "$restored_volume" >/dev/null
restore_started=$(date +%s)
docker run --rm "${walg_env[@]}" --network "$network" --user root -v "$restored_volume:/var/lib/postgresql/data" \
  --entrypoint bash "$image" -c '
    set -e
    wal-g backup-list
    wal-g backup-fetch /var/lib/postgresql/data LATEST
    touch /var/lib/postgresql/data/recovery.signal
    chown -R postgres:postgres /var/lib/postgresql/data
    chmod 700 /var/lib/postgresql/data'

docker run -d --name "$restored_db" --network "$network" "${walg_env[@]}" \
  -v "$restored_volume:/var/lib/postgresql/data" "$image" \
  postgres -c "restore_command=wal-g wal-fetch %f %p" \
  -c "recovery_target_time=$target_time" -c recovery_target_action=promote >/dev/null
wait_ready "$restored_db"
for _ in $(seq 1 60); do
  [ "$(sql "$restored_db" "select pg_is_in_recovery()")" = "f" ] && break
  sleep 1
done
restore_seconds=$(elapsed_since "$restore_started")

total=$(sql "$restored_db" "select count(*) from ledger")
before=$(sql "$restored_db" "select count(*) from ledger where batch = 'before-backup'")
after=$(sql "$restored_db" "select count(*) from ledger where batch = 'after-backup'")
past_target=$(sql "$restored_db" "select count(*) from ledger where batch = 'after-target'")
echo "restored rows: total=$total before-backup=$before after-backup=$after after-target=$past_target in ${restore_seconds}s"

[ "$before" = "1000" ] || { echo "expected 1000 rows from before the base backup, got $before"; exit 1; }
[ "$after" = "500" ] || { echo "expected 500 rows replayed from archived WAL, got $after"; exit 1; }
[ "$past_target" = "0" ] || { echo "found $past_target rows written after the recovery target"; exit 1; }
sql "$restored_db" "insert into ledger (batch, amount) values ('post-restore', 1)"
echo "restore drill passed: point-in-time recovery stopped at $target_time"
