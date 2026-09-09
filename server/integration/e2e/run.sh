#!/usr/bin/env bash
set -euo pipefail

integration_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
image="${NEKO_E2E_IMAGE:-neko-m1-e2e:local}"
base_url="${NEKO_E2E_BASE_URL:-http://127.0.0.1:8080}"

: "${NEKO_E2E_PASSWORD:?set NEKO_E2E_PASSWORD}"

if ! docker image inspect "${image}" >/dev/null 2>&1; then
  echo "==> building browser E2E image ${image}"
  docker build --tag "${image}" "${integration_dir}"
fi

args=(
  --rm
  --network host
  -e "NEKO_E2E_BASE_URL=${base_url}"
  -e "NEKO_E2E_USERNAME=${NEKO_E2E_USERNAME:-e2e-browser}"
  -e "NEKO_E2E_PASSWORD=${NEKO_E2E_PASSWORD}"
  -e "NEKO_E2E_TIMEOUT_MS=${NEKO_E2E_TIMEOUT_MS:-45000}"
  -e "NEKO_E2E_HEADLESS=${NEKO_E2E_HEADLESS:-1}"
  -e "NEKO_E2E_PROFILE=${NEKO_E2E_PROFILE:-unspecified}"
  -v "${integration_dir}:/tests:ro"
)

if [[ -n "${NEKO_E2E_OUTPUT:-}" ]]; then
  output_dir="$(dirname "${NEKO_E2E_OUTPUT}")"
  mkdir -p "${output_dir}"
  args+=(
    -e "NEKO_E2E_OUTPUT=/output/$(basename "${NEKO_E2E_OUTPUT}")"
    -v "${output_dir}:/output"
  )
fi

if [[ -n "${NEKO_E2E_ARTIFACT_DIR:-}" ]]; then
  mkdir -p "${NEKO_E2E_ARTIFACT_DIR}"
  args+=(
    -e "NEKO_E2E_ARTIFACT_DIR=/artifacts"
    -v "${NEKO_E2E_ARTIFACT_DIR}:/artifacts"
  )
fi

docker run "${args[@]}" "${image}"
