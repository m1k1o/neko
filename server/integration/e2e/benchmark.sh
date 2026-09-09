#!/usr/bin/env bash
set -euo pipefail

integration_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
output_dir="${NEKO_E2E_OUTPUT_DIR:-${integration_dir}/results}"
viewers="${NEKO_E2E_VIEWERS:-1}"
user_prefix="${NEKO_E2E_USERNAME_PREFIX:-e2e-viewer}"
image="${NEKO_E2E_IMAGE:-neko-m1-e2e:local}"

: "${NEKO_E2E_PASSWORD:?set NEKO_E2E_PASSWORD}"
mkdir -p "${output_dir}"

if ! docker image inspect "${image}" >/dev/null 2>&1; then
  echo "==> building browser E2E image ${image}"
  docker build --tag "${image}" "${integration_dir}"
fi

echo "==> browser baseline: ${viewers} viewer(s), profile=${NEKO_E2E_PROFILE:-unspecified}"
pids=()
for viewer in $(seq 1 "${viewers}"); do
  output="${output_dir}/viewer-${viewer}.json"
  NEKO_E2E_USERNAME="${user_prefix}-${viewer}" \
    NEKO_E2E_IMAGE="${image}" \
    NEKO_E2E_SKIP_BUILD=1 \
    NEKO_E2E_OUTPUT="${output}" \
    NEKO_E2E_COLLECT_METRICS="${NEKO_E2E_COLLECT_METRICS:-0}" \
    NEKO_E2E_METRICS_OUTPUT="${output_dir}/viewer-${viewer}.prom" \
    NEKO_E2E_ARTIFACT_DIR="${output_dir}/viewer-${viewer}-artifacts" \
    "${integration_dir}/run.sh" >"${output_dir}/viewer-${viewer}.log" 2>&1 &
  pids+=("$!")
done

status=0
for pid in "${pids[@]}"; do
  if ! wait "${pid}"; then
    status=1
  fi
done

if [[ "${status}" -ne 0 ]]; then
  echo "browser baseline failed; inspect ${output_dir}/*.log" >&2
  exit "${status}"
fi

echo "browser baseline results: ${output_dir}"
