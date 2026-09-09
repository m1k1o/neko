# Local Chromium M1 demo

This demo overlays locally built M1 binaries and Chromium integration files on
the published Neko Chromium runtime. It does not store passwords in Git.
It uses the explicit `balanced` profile (`1280x720@30`, 2500 kbit/s) so the
standard M1 quality pipeline is exercised on every demo start.

Build the current branch binaries:

```bash
cd server
go build -o bin/neko ./cmd/neko
CGO_ENABLED=0 go build -o bin/neko-proxy ./cmd/neko-proxy
cd ..
```

Set credentials and network addresses. For native Linux, the loopback defaults
are sufficient. For a Windows browser connecting to Neko inside WSL2, use the
first WSL address for the media bind and advertised NAT address:

```bash
export NEKO_DEMO_USER_PASSWORD='replace-me'
export NEKO_DEMO_ADMIN_PASSWORD='replace-me-too'
export NEKO_DEMO_MEDIA_BIND='127.0.0.1'
export NEKO_DEMO_NAT_IP='127.0.0.1'
```

Then build and start the demo:

```bash
docker compose -f demo/compose.local.yaml up -d --build
```

For a no-public-IP deployment, copy `demo/compose.frp.example.yaml` to
`compose.frp.yaml`, copy `demo/frpc.toml.example` to `frpc.toml`, and replace
the FRP/SakuraFrp placeholders. The FRP server must expose both TCP and UDP
proxies on the same `${FRP_MEDIA_PORT}` (default `52000`); set
`FRP_PUBLIC_IP` to the FRP node's public address, not the local host address.

For TURN fallback, copy `demo/compose.turn.example.yaml` to
`compose.turn.yaml` and set the TURN host, public IP, relay range, and
credentials. Open TCP/UDP `3478` plus the complete relay range in the TURN
server firewall. Do not commit the copied files containing credentials.

Open <http://127.0.0.1:8080>. Stop it with:

```bash
docker compose -f demo/compose.local.yaml down
```
