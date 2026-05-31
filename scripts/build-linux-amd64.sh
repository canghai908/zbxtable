#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$(cd "${ROOT_DIR}/../zbxtable-web" && pwd)"
EMBED_DIR="${ROOT_DIR}/api/v1/web/web"
APP_NAME="zbxtable"
DIST_DIR="${ROOT_DIR}/dist"
OUTPUT="${DIST_DIR}/${APP_NAME}-linux-amd64"

VERSION="$(git -C "${ROOT_DIR}" describe --tags --exact-match 2>/dev/null || echo "dev")"
GIT_HASH="$(git -C "${ROOT_DIR}" rev-parse --short HEAD 2>/dev/null || echo "unknown")"
BUILD_TIME="$(date -u '+%Y-%m-%d_%H:%M:%S')"
LDFLAGS="-X main.version=${VERSION} -X main.gitHash=${GIT_HASH} -X main.buildTime=${BUILD_TIME} -w -s"

mkdir -p "${DIST_DIR}"

echo "Building frontend bundle..."
(
  cd "${FRONTEND_DIR}"
  yarn build
)

echo "Syncing frontend bundle into embedded web directory..."
rm -rf "${EMBED_DIR}"
mkdir -p "${EMBED_DIR}"
cp -R "${FRONTEND_DIR}/web/." "${EMBED_DIR}/"

echo "Building ${APP_NAME} for linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags "${LDFLAGS}" -o "${OUTPUT}" "${ROOT_DIR}/main.go"

echo "Build complete: ${OUTPUT}"
