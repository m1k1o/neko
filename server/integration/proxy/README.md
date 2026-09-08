# Chromium proxy integration test

This test starts an HTTP origin, an authenticated Squid HTTP proxy, an
authenticated microsocks SOCKS5 proxy, and one Neko proxy agent for each
upstream protocol.

Run it from the repository root with Docker Engine and Docker Compose v2:

```bash
server/integration/proxy/run.sh
```

The script verifies active health checks, regular HTTP forwarding, CONNECT
tunnels, invalid credentials, and credential-free diagnostics. It removes all
test containers and volumes when it exits.
