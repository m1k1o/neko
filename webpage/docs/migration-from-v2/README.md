---
description: Migrating clients and configuration to the M1 runtime.
---

# M1 migration

The M1 runtime has one protocol and one configuration path. The old V2/legacy
HTTP API, WebSocket endpoint, client-created data channel, flat WebSocket
messages, and compatibility configuration flags have been removed.

## Client and API changes

- Serve the client from the normal application path and connect to `/api/ws`.
- Authenticate with `POST /api/login` using `{ "username", "password" }`.
  The server returns a token and also establishes the session cookie.
- Use the canonical WebSocket envelope `{ "event": "...", "payload": {} }`.
  Events without data may omit `payload`; flat event objects are rejected.
- Let the server create the WebRTC data channel. Clients must only accept the
  negotiated `data` channel.

Upgrade the server and client together. An old client will not be silently
adapted; it must be upgraded before connecting to an M1 server.

## Configuration changes

Use the namespaced options documented in the current configuration reference:

- `filetransfer.enabled` and `filetransfer.dir` replace the removed
  `file_transfer_enabled` and `file_transfer_path` options.
- Configure WebRTC connectivity through the current `webrtc.*` options. The
  old `legacy` switch and V2 pipeline settings no longer exist.
- Keep the deployment ports explicit: HTTP(S) and the configured WebRTC MUX
  port (or the documented FRP/TURN fallback).

Copy the old configuration, remove unsupported keys, and validate it with
`neko serve --help` before starting the upgraded server. Do not mix old and
new client bundles in the same deployment.
