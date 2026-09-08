#!/usr/bin/env bash
set -euo pipefail

launcher_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

output="$(
  NEKO_CHROMIUM_BIN=/bin/echo \
  NEKO_CHROMIUM_PROXY_AGENT=127.0.0.1:18080 \
  NEKO_CHROMIUM_PROXY_BYPASS_LIST='localhost;127.0.0.1;*.internal' \
  "${launcher_dir}/neko-chromium" --no-sandbox
)"

expected='--no-sandbox --proxy-server=http://127.0.0.1:18080 --proxy-bypass-list=localhost;127.0.0.1;*.internal'
if [[ "${output}" != "${expected}" ]]; then
  echo "unexpected Chromium arguments: ${output}" >&2
  exit 1
fi

if NEKO_CHROMIUM_BIN=/bin/true NEKO_CHROMIUM_PROXY_AGENT=proxy.example.test:8080 "${launcher_dir}/neko-chromium"; then
  echo "non-loopback agent address unexpectedly succeeded" >&2
  exit 1
fi
