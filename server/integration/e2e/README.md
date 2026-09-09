# Chromium WebRTC E2E and baseline

This suite drives the built client with a real Chromium browser. It checks the
authenticated login flow, the canonical WebSocket envelope, ICE connection
state, and receipt of a non-empty remote video frame. The result is JSON so it
can be archived by CI or compared across 720p/1080p and encoder profiles.

Start a Neko instance first (the local demo is the supported smoke target):

```bash
export NEKO_DEMO_USER_PASSWORD='replace-me'
export NEKO_DEMO_ADMIN_PASSWORD='replace-me-too'
docker compose -f demo/compose.local.yaml up -d --build
```

Run one browser check. The first run builds a pinned Playwright image and may
download the browser runtime:

```bash
export NEKO_E2E_PASSWORD="$NEKO_DEMO_ADMIN_PASSWORD"
export NEKO_E2E_OUTPUT=/tmp/neko-e2e.json
server/integration/e2e/run.sh
```

The output records connection/first-frame latency, video dimensions, received
signaling events, browser version, and the selected profile. It never writes
the password to output or logs. Set `NEKO_E2E_ARTIFACT_DIR` to save a failure
screenshot. `NEKO_E2E_BASE_URL` changes the target; the runner uses host
networking so WebRTC's media port is exercised as well as HTTP/WebSocket.

To run concurrent viewer samples (for example, 1, 2, and 5 viewers), set the
count and repeat the command with the server's desired capture profile:

```bash
export NEKO_E2E_VIEWERS=5
export NEKO_E2E_PROFILE=balanced-720p
server/integration/e2e/benchmark.sh
```

Results are written under `server/integration/e2e/results` by default. This is
an explicit local artifact directory and is ignored by Git; remove it after
review if it is no longer needed.
