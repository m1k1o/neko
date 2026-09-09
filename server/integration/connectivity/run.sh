#!/usr/bin/env bash
set -euo pipefail

integration_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
active_suite=""

cleanup() {
  if [[ -n "${active_suite}" ]]; then
    (cd "${integration_dir}/${active_suite}" && docker compose down --volumes --remove-orphans)
  fi
}
trap cleanup EXIT

run_suite() {
  local suite="$1"
  active_suite="${suite}"
  cd "${integration_dir}/${suite}"

  echo "==> ${suite} connectivity integration"
  if [[ "${suite}" == "frp" ]]; then
    docker compose up --abort-on-container-exit --exit-code-from probe
  else
    TURN_TRANSPORT=udp docker compose up --abort-on-container-exit --exit-code-from probe
    docker compose down --volumes --remove-orphans
    TURN_TRANSPORT=tcp docker compose up --abort-on-container-exit --exit-code-from probe
  fi
  docker compose down --volumes --remove-orphans
  active_suite=""
}

run_suite frp
run_suite turn
