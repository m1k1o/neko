#!/usr/bin/env bash
set -euo pipefail

integration_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${integration_dir}"

cleanup() {
  docker compose down --volumes --remove-orphans
}
trap cleanup EXIT

docker compose up --detach --build --wait origin squid socks http-agent socks-agent

for agent in http-agent socks-agent; do
  health="$(docker compose exec -T "${agent}" curl -fsS http://127.0.0.1:18081/healthz)"
  if [[ "${health}" != *'"status":"healthy"'* ]]; then
    echo "${agent} returned an unexpected health response: ${health}" >&2
    exit 1
  fi
  if [[ "${health}" == *'proxy-password'* || "${health}" == *'proxy-user'* ]]; then
    echo "${agent} leaked proxy credentials in its health response" >&2
    exit 1
  fi

  body="$(docker compose exec -T "${agent}" curl -fsS --noproxy '' --proxy http://127.0.0.1:18080 http://origin:8080/)"
  if [[ "${body}" != *'Directory listing for /'* ]]; then
    echo "${agent} did not forward the HTTP request to the origin" >&2
    exit 1
  fi

  tunnel_body="$(docker compose exec -T "${agent}" curl -fsS --noproxy '' --proxytunnel --proxy http://127.0.0.1:18080 http://origin:8080/)"
  if [[ "${tunnel_body}" != *'Directory listing for /'* ]]; then
    echo "${agent} did not establish a CONNECT tunnel to the origin" >&2
    exit 1
  fi
done

for agent in http-agent socks-agent; do
  if bad_output="$(docker compose run --rm \
    --env NEKO_CHROMIUM_PROXY_USERNAME=wrong-user \
    "${agent}" 2>&1)"; then
    echo "${agent} unexpectedly started with invalid credentials" >&2
    exit 1
  fi
  if [[ "${bad_output}" != *'authentication_failed'* ]]; then
    echo "${agent} did not report an authentication failure" >&2
    exit 1
  fi
  if [[ "${bad_output}" == *'proxy-password'* ]]; then
    echo "${agent} leaked the proxy password in startup diagnostics" >&2
    exit 1
  fi
done
