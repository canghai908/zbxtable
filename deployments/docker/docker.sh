#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.yml"

MYSQL_CONTAINER="${MYSQL_CONTAINER:-zbxtable-mysql}"
MYSQL_USER="${MYSQL_USER:-zbxtable}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-zbxtablepwd123}"
MYSQL_DATABASE="${MYSQL_DATABASE:-zbxtable}"

compose() {
  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose -f "${COMPOSE_FILE}" "$@"
  else
    docker compose -f "${COMPOSE_FILE}" "$@"
  fi
}

usage() {
  cat <<'EOF'
ZbxTable Docker 管理脚本

用法:
  ./docker.sh <命令> [参数]

命令:
  up                    启动服务（mysql + zbxtable，自动创建运行目录）
  down                  停止并移除容器
  restart               重启服务
  ps                    查看服务状态
  logs [service]        查看日志，默认查看全部服务
  build                 重新构建镜像
  rebuild               强制重建并启动（自动创建运行目录 + down + build --no-cache + up -d）
  pull                  拉取基础镜像
  backup [文件路径]      备份数据库，默认输出到当前目录 backup_时间.sql
  restore <文件路径>     从 SQL 文件恢复数据库
  help                  查看帮助

示例:
  ./docker.sh up
  ./docker.sh logs zbxtable
  ./docker.sh backup ./backup/zbxtable_$(date +%F).sql
  ./docker.sh restore ./backup/zbxtable_2026-03-06.sql
EOF
}

ensure_compose_file() {
  if [[ ! -f "${COMPOSE_FILE}" ]]; then
    echo "[错误] 未找到 docker-compose.yml: ${COMPOSE_FILE}" >&2
    exit 1
  fi
}

ensure_runtime_dirs() {
  mkdir -p "${PROJECT_ROOT}/config" "${PROJECT_ROOT}/log" "${PROJECT_ROOT}/download" "${PROJECT_ROOT}/template"
}

backup_db() {
  local output="${1:-backup_$(date +%Y%m%d_%H%M%S).sql}"
  mkdir -p "$(dirname "${output}")"
  docker exec "${MYSQL_CONTAINER}" \
    sh -c "exec mysqldump -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}" > "${output}"
  echo "[完成] 数据库备份成功: ${output}"
}

restore_db() {
  local input="${1:-}"
  if [[ -z "${input}" ]]; then
    echo "[错误] 请提供恢复文件路径" >&2
    echo "示例: ./docker.sh restore ./backup/zbxtable.sql" >&2
    exit 1
  fi

  if [[ ! -f "${input}" ]]; then
    echo "[错误] 恢复文件不存在: ${input}" >&2
    exit 1
  fi

  docker exec -i "${MYSQL_CONTAINER}" \
    sh -c "exec mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}" < "${input}"
  echo "[完成] 数据库恢复成功: ${input}"
}

main() {
  local cmd="${1:-help}"
  shift || true

  ensure_compose_file

  case "${cmd}" in
    up)
      ensure_runtime_dirs
      compose up -d
      ;;
    down)
      compose down
      ;;
    restart)
      compose restart
      ;;
    ps)
      compose ps
      ;;
    logs)
      if [[ $# -gt 0 ]]; then
        compose logs -f "$1"
      else
        compose logs -f
      fi
      ;;
    build)
      compose build
      ;;
    rebuild)
      ensure_runtime_dirs
      compose down
      compose build --no-cache
      compose up -d
      ;;
    pull)
      compose pull
      ;;
    backup)
      backup_db "$@"
      ;;
    restore)
      restore_db "$@"
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      echo "[错误] 未知命令: ${cmd}" >&2
      usage
      exit 1
      ;;
  esac
}

main "$@"
