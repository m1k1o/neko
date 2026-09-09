# FRP and TURN connectivity integration tests

These tests exercise the two constrained-network deployment paths without
requiring a public server:

- `frp/` starts a local FRP server and client, then verifies that the same
  public port forwards both TCP and UDP traffic to an internal endpoint. This
  models a host with no public IP.
- `turn/` starts Coturn with long-term credentials and a bounded relay range,
  then performs authenticated STUN Allocate requests over UDP and TCP. The
  returned relay port is checked against the configured range.

Run both suites from the repository root with Docker Engine and Compose v2:

```bash
server/integration/connectivity/run.sh
```

The script removes all test containers, networks, and volumes when it exits.
The images are pinned to the same FRP and Coturn versions used by the M1
deployment templates. No Neko credentials or public addresses are required.
