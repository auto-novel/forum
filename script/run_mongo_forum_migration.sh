#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd -- "$script_dir/.." && pwd)"
auth_project_root="${AUTH_PROJECT_ROOT:-$(cd -- "$project_root/../auth" && pwd)}"

forum_container="${FORUM_DB_CONTAINER:-forum-postgresql-1}"
auth_container="${AUTH_DB_CONTAINER:-auth-postgresql-1}"
mongo_container="${MONGO_CONTAINER:-auto-novel-mongo-1}"
forum_network="${FORUM_NETWORK:-forum}"
auth_network="${AUTH_NETWORK:-auth_default}"
mongo_network="${MONGO_NETWORK:-auto-novel}"
migrator_image="${MIGRATOR_IMAGE:-python:3.12-slim}"
migrator_container="forum-migrator-$$"
output_dir="${MIGRATION_OUTPUT_DIR:-$project_root/data/migration}"
migrator_started=false

usage() {
  cat <<'EOF'
用法：script/run_mongo_forum_migration.sh [迁移参数]

默认只执行预检；传入 --execute 后才会写入 forum 数据库。

示例：
  script/run_mongo_forum_migration.sh
  script/run_mongo_forum_migration.sh --execute
  script/run_mongo_forum_migration.sh --execute --mapping-file /output/custom-map.jsonl

可通过环境变量覆盖容器名、网络和输出目录：
  FORUM_DB_CONTAINER  AUTH_DB_CONTAINER  MONGO_CONTAINER
  FORUM_NETWORK       AUTH_NETWORK       MONGO_NETWORK
  MIGRATOR_IMAGE      MIGRATION_OUTPUT_DIR AUTH_PROJECT_ROOT
EOF
}

cleanup() {
  if [[ "$migrator_started" == true ]]; then
    docker stop "$migrator_container" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT INT TERM

require_file() {
  if [[ ! -f "$1" ]]; then
    echo "缺少文件：$1" >&2
    exit 1
  fi
}

require_container() {
  if ! docker inspect "$1" >/dev/null 2>&1; then
    echo "找不到 Docker 容器：$1" >&2
    exit 1
  fi
}

wait_until_ready() {
  local container="$1"
  local status
  for _ in {1..60}; do
    status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container")"
    case "$status" in
      healthy | running)
        return 0
        ;;
      unhealthy | exited | dead)
        echo "容器 $container 未就绪，当前状态：$status" >&2
        docker logs --tail 50 "$container" >&2 || true
        return 1
        ;;
    esac
    sleep 1
  done
  echo "等待容器 $container 就绪超时" >&2
  return 1
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_file "$project_root/script/migrate_mongo_forum.py"
require_file "$project_root/script/requirements-migration.txt"
require_file "$project_root/.env"
require_file "$auth_project_root/.env"

require_container "$forum_container"
require_container "$auth_container"
require_container "$mongo_container"

for container in "$forum_container" "$auth_container" "$mongo_container"; do
  if [[ "$(docker inspect --format '{{.State.Running}}' "$container")" != true ]]; then
    echo "启动数据库容器：$container"
    docker start "$container" >/dev/null
  fi
  wait_until_ready "$container"
done

mkdir -p "$output_dir"

echo "创建迁移容器：$migrator_container"
docker run --detach --rm \
  --name "$migrator_container" \
  --network "$forum_network" \
  --volume "$project_root:/work:ro" \
  --volume "$project_root/.env:/run/secrets/forum.env:ro" \
  --volume "$auth_project_root/.env:/run/secrets/auth.env:ro" \
  --volume "$output_dir:/output" \
  "$migrator_image" sleep infinity >/dev/null
migrator_started=true

docker network connect "$mongo_network" "$migrator_container"
docker network connect "$auth_network" "$migrator_container"

migration_args=("$@")
has_execute=false
has_mapping_file=false
for argument in "${migration_args[@]}"; do
  case "$argument" in
    --execute)
      has_execute=true
      ;;
    --mapping-file | --mapping-file=*)
      has_mapping_file=true
      ;;
  esac
done

if [[ "$has_execute" == true && "$has_mapping_file" == false ]]; then
  mapping_name="id-map-$(date -u +%Y%m%dT%H%M%SZ)-$$.jsonl"
  migration_args+=(--mapping-file "/output/$mapping_name")
  echo "ID 映射将保存到：$output_dir/$mapping_name"
fi

docker exec "$migrator_container" \
  sh -c '
    set -eu

    . /run/secrets/forum.env
    forum_password=$POSTGRES_PASSWORD
    . /run/secrets/auth.env
    auth_password=$POSTGRES_PASSWORD

    export MONGO_URI="mongodb://'$mongo_container':27017"
    export MONGO_DB=main
    export FORUM_DATABASE_URL="postgresql://forum:${forum_password}@'$forum_container':5432/forum"
    export AUTH_DATABASE_URL="postgresql://auth:${auth_password}@'$auth_container':5432/auth"

    python -m pip install --disable-pip-version-check \
      --requirement /work/script/requirements-migration.txt
    exec python /work/script/migrate_mongo_forum.py "$@"
  ' migration-runner "${migration_args[@]}"
