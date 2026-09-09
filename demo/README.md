# Local Chromium M1 demo

This demo overlays locally built M1 binaries and Chromium integration files on
the published Neko Chromium runtime. It does not store passwords in Git.

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

Open <http://127.0.0.1:8080>. Stop it with:

```bash
docker compose -f demo/compose.local.yaml down
```
