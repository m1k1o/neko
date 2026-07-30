---
name: dev-guide
description: Development guide for the neko project. Use when asked to add features, fix bugs, or make changes to the server (Go) or client (Vue 2) codebase. Covers architecture, conventions, build commands, and common patterns.
---

# Neko Development Guide

## Project Overview

Neko is a self-hosted virtual browser/desktop sharing system. It streams a Docker container's X11 desktop to multiple users via WebRTC, with bidirectional mouse/keyboard control through WebRTC DataChannels.

## Repository Structure

```
neko/
├── server/          # Go backend
│   ├── cmd/         # CLI entrypoints (cobra)
│   ├── internal/    # Private packages
│   │   ├── api/     # HTTP REST API (chi router)
│   │   ├── capture/ # GStreamer screen capture
│   │   ├── desktop/ # X11/Xorg input injection
│   │   ├── http/    # HTTP server & WebSocket
│   │   ├── member/  # Auth providers
│   │   ├── plugins/ # Plugin system
│   │   ├── session/ # Session management
│   │   └── webrtc/  # WebRTC (pion)
│   └── pkg/types/   # Shared interfaces
├── client/          # Vue 2 frontend
│   └── src/
│       ├── components/
│       ├── store/   # Vuex modules
│       └── api/     # Axios API client
├── apps/            # Per-browser Docker configs
│   ├── firefox/
│   ├── chromium/
│   └── ...
└── webpage/         # Docusaurus docs site
```

## Tech Stack

### Backend (Go 1.25)
- **Router**: go-chi/chi v5
- **WebRTC**: pion/webrtc v4 (pure Go)
- **WebSocket**: gorilla/websocket
- **Config**: spf13/viper + cobra
- **Logging**: rs/zerolog (structured, use `log.Info().Str("key", val).Msg("...")`)
- **Metrics**: prometheus/client_golang
- **Desktop**: X11/Xorg via CGo bindings
- **Capture**: GStreamer via CGo bindings

### Frontend (TypeScript + Vue 2)
- **Framework**: Vue 2.7 with Class-style components (`vue-class-component` + `vue-property-decorator`)
- **State**: Vuex 3 + typed-vuex
- **Build**: Vue CLI 5 (webpack)
- **Styles**: SCSS

## Build Commands

### Server
```bash
cd server
go build ./...          # build
go test ./...           # run tests
./dev/build             # dev build script
./dev/start             # start dev server
./dev/fmt               # format (gofmt)
./dev/lint              # lint
```

### Client
```bash
cd client
npm install
npm run serve           # dev server with HMR
npm run build           # production build
npm run lint            # ESLint
```

### Docker (full stack)
```bash
docker compose up       # uses docker-compose.yaml at root
```

## Server Conventions

### Adding a new API endpoint

1. Add handler method to the relevant controller in `server/internal/api/`
2. Register the route in `server/internal/api/router.go`
3. Follow the existing pattern — handlers receive `(w http.ResponseWriter, r *http.Request)` and use chi's context for path params

```go
// Example handler pattern
func (h *RoomHandler) screenGet(w http.ResponseWriter, r *http.Request) {
    session := h.sessions.GetSession(r)
    // ... logic
    utils.HttpSuccess(w, response)
}
```

### Adding a new plugin

Plugins live in `server/internal/plugins/<name>/`. Each plugin implements the `types.Plugin` interface:

```go
type Plugin interface {
    Name() string
    Start() error
    Stop() error
}
```

Register in `server/cmd/plugins.go`.

### Configuration

All config is in `server/internal/config/`. Each subsystem has its own config struct with viper bindings. Environment variables follow the pattern `NEKO_<SUBSYSTEM>_<FIELD>` (e.g., `NEKO_WEBRTC_EPR`).

### Logging

Use zerolog — never `fmt.Println` or `log.Printf`:

```go
h.logger.Info().Str("session_id", session.ID()).Msg("session connected")
h.logger.Error().Err(err).Msg("failed to process event")
```

## Client Conventions

### Component style

Use Class-style components with decorators:

```typescript
@Component({ components: { MyChild } })
export default class MyComponent extends Vue {
  @Prop({ required: true }) readonly value!: string
  @Watch('value') onValueChange(val: string) { ... }

  get computed() { return this.$store.state.something }
}
```

### Vuex store

Modules live in `client/src/store/`. Use `typed-vuex` accessors — avoid direct `this.$store.commit()` calls; use the typed accessor instead.

### API calls

Use the existing axios client in `client/src/api/`. Don't create new axios instances.

### Styles

- SCSS only, no plain CSS files
- Component-scoped styles with `<style lang="scss" scoped>`
- Global variables/mixins are in `client/src/assets/styles/`

## WebRTC Data Flow

```
GStreamer (X11 capture)
    → pion/webrtc encoder
    → WebRTC video track → browser client

Browser client (mouse/keyboard events)
    → WebRTC DataChannel (binary, big-endian)
    → server/internal/webrtc/handler.go
    → server/internal/desktop/xinput.go (X11 injection)
```

DataChannel messages use a binary header defined in `server/internal/webrtc/payload/`. Always use `binary.BigEndian` for encoding/decoding.

## Testing

### Server tests
```bash
cd server && go test ./...
```

Tests use real structs, not mocks where possible. Integration tests for member providers are in `server/internal/member/file/provider_test.go`.

### Client tests
```bash
cd client && npm test
```

## Common Pitfalls

- **CGo required**: The server uses CGo for X11 and GStreamer. Cross-compiling needs the right CGo toolchain. Use the provided Docker dev environment (`server/dev/runtime/`) for Linux builds.
- **WebRTC port range**: UDP ports `52000-52100` must be exposed in Docker. Set `NEKO_WEBRTC_EPR=52000-52100`.
- **ICE/NAT**: For external access, set `NEKO_NAT1TO1=<public_ip>` or configure a TURN server.
- **Shared memory**: Docker needs `shm_size: 2gb` for browser rendering.
- **Plugin dependency order**: Plugins declare dependencies via `server/internal/plugins/dependency.go`. Circular deps will panic at startup.
