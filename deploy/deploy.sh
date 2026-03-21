#!/usr/bin/env bash

set -euo pipefail

APP_DIR="${DEPLOY_PATH:-/srv/legend-portal/app}"
SHARED_DIR="${SHARED_PATH:-/srv/legend-portal/shared}"
CONFIG_TEMPLATE="${APP_DIR}/configs/config.production.yaml"
CONFIG_PATH="${APP_CONFIG_PATH:-${SHARED_DIR}/config.production.yaml}"
SERVICE_NAME="${SERVICE_NAME:-legend-portal}"

mkdir -p "${SHARED_DIR}/data" "${SHARED_DIR}/uploads" "${SHARED_DIR}/logs"

cd "${APP_DIR}"

git fetch origin main
git reset --hard origin/main

if [ ! -f "${CONFIG_PATH}" ]; then
  cp "${CONFIG_TEMPLATE}" "${CONFIG_PATH}"
  echo "created config template at ${CONFIG_PATH}"
fi

export APP_CONFIG_PATH="${CONFIG_PATH}"
export GOCACHE="${GOCACHE:-/tmp/go-build}"
export GOMODCACHE="${GOMODCACHE:-/tmp/go-mod-cache}"

go build -o legend-portal ./cmd/web

systemctl restart "${SERVICE_NAME}"
systemctl is-active --quiet "${SERVICE_NAME}"

curl --fail --silent http://127.0.0.1:8080/healthz >/dev/null

echo "deploy completed"
