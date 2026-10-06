# neko client

The web client for [neko](https://github.com/m1k1o/neko): React 19 + TypeScript, built with Vite.
It speaks only the v3 API (`/api/ws` + REST), so it works with `NEKO_LEGACY=false`.

```
src/
  core/   framework-free client library: no React, no runtime dependencies
  app/    the GUI (React), reusing the look of the previous Vue 2 client
```

## core

`src/core` is everything needed to talk to a neko server, usable from any framework:

- `client.ts`: `NekoClient`: auth (`login`, `logout`, `setUrl`, `?token=`), websocket with
  reconnect + stale detection, WebRTC (server offers, client answers; early ICE candidates are
  buffered), the binary input data channel, control/screen/clipboard/broadcast messages.
- `overlay.ts`: input layer over the video (mouse, wheel, keyboard via the vendored Apache Guacamole
  keyboard, touch, file drop) and drawing the host's cursor for everyone else.
- `store.ts`: tiny proxy store (`subscribe`, `watch`, `version`) + typed event emitter.
- `api.ts`: fetch wrapper for the REST API.
- `types.ts`: client state and wire types, mirroring `server/pkg/types`.

```ts
const client = new NekoClient({ autologin: true, autoconnect: true, autoplay: true })
client.store.subscribe(() => render(client.state)) // state: connection, sessions, control, screen, ...
client.events.on('message', (event, payload) => {}) // plugin events (chat/*, filetransfer/*, ...)
client.setUrl('https://neko.example.com/')
client.mount(document.getElementById('video')!)
```

## Development

Backend from `server/dev`, then:

```sh
cd client/dev && ./serve        # http://localhost:3001, /api proxied to the backend on :3000
# or without docker:
NEKO_URL=http://localhost:3000 npm run dev
```

| script                |                                                            |
| --------------------- | ---------------------------------------------------------- |
| `npm run build`       | type-check + production build into `dist/`                 |
| `npm run check`       | types, formatting, and the `*.check.ts` self-checks        |
| `npm run format`      | prettier                                                   |
| `npm run build:emoji` | regenerate emoji data (`public/emoji.json`, `_emoji.scss`) |
| `npm run test:e2e`    | two-user browser test against a running server, see below  |

### e2e

`e2e.mjs` drives two (or more) headless Chromium users through login, video, chat, control, files,
admin actions, reconnects and URL parameters. It needs a disposable server (it takes control, types
into the desktop and kicks users) with the legacy API off and the multiuser provider:

```sh
docker run -d -p 8080:8080 -p 52000:52000/udp -e NEKO_LEGACY=false \
  -e NEKO_MEMBER_PROVIDER=multiuser -e NEKO_MEMBER_MULTIUSER_USER_PASSWORD=neko \
  -e NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD=admin -e NEKO_FILETRANSFER_ENABLED=true \
  -e NEKO_WEBRTC_UDPMUX=52000 \
  -e NEKO_WEBRTC_NAT1TO1=127.0.0.1 ghcr.io/m1k1o/neko/kde
npx playwright install chromium
NEKO_URL=http://localhost:8080/ npm run test:e2e
```

The typing step expects KRunner on Alt+F2 (KDE). Set `NEKO_FILE_URL` to a second server using the
`file` member provider (members `alice`/`alice` admin, `bob`/`bob`) to also test ban/unban.

## Behaviour notes

- **Kick / ban** follow common chat-app semantics, see the comment at `actions.kick` in
  `src/app/neko.ts`: kick removes the session (the user may log in again), ban sets
  `can_login: false` and is only offered when the auth provider stores accounts.
- Settings use the same `localStorage` keys and format as the previous client, so preferences carry
  over. Scroll sensitivity now uses v3 steps (-5..5, key `scroll_sensitivity`); `?scroll=` is mapped.
- Chat markdown is parsed to React elements (no HTML strings); links are limited to http(s)/mailto.
- File downloads are plain links so the browser streams them; the session token is added to the
  URL only when the server runs without cookies.
