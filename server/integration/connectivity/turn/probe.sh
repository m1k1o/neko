#!/usr/bin/env sh
set -eu

transport="${TURN_TRANSPORT:-udp}"
case "${transport}" in
  udp) client_args="" ;;
  tcp) client_args="-t" ;;
  *) echo "unsupported TURN transport: ${transport}" >&2; exit 2 ;;
esac

set +e
output="$(turnutils_uclient coturn -u turn-user -w turn-password -p 3478 -y -n 1 -c -v ${client_args} 2>&1)"
status=$?
set -e
printf '%s\n' "${output}"
if [ "${status}" -ne 0 ]; then
  echo "TURN ${transport} allocation failed (check relay reachability and credentials)" >&2
  exit "${status}"
fi

relay_ports="$(printf '%s\n' "${output}" | sed -n 's/.*Received relay addr: .*:\([0-9][0-9]*\)$/\1/p')"
if [ -z "${relay_ports}" ]; then
  echo "TURN ${transport} allocation returned no relay address" >&2
  exit 1
fi
for relay_port in ${relay_ports}; do
  if [ "${relay_port}" -lt 49160 ] || [ "${relay_port}" -gt 49170 ]; then
    echo "TURN ${transport} relay port ${relay_port} is outside 49160-49170" >&2
    exit 1
  fi
done
echo "TURN ${transport} 3478: authenticated, relay ports are within 49160-49170"
