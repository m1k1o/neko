# @m1k1o/neko

Framework-free client library for the [neko](https://github.com/m1k1o/neko) v3 API: REST, the
`/api/ws` event protocol, the stream through a pluggable transport (WebRTC with its binary input
channel, see [Transports](#transports)), and the input overlay (mouse, wheel, keyboard, touch, file
drop, the host's cursor). It is what the React GUI in `../src` is built on, and it can drive a neko
server from any framework or from a plain page.

ESM, TypeScript declarations included, no runtime dependencies.

```sh
npm install @m1k1o/neko
```

## Quick start

```ts
import { NekoClient } from '@m1k1o/neko'

const client = new NekoClient({ autologin: true, autoconnect: true, autoplay: true })

// state: connection, video, control, screen, sessions, settings
client.store.subscribe(() => render(client.state))
// one-off happenings, including plugin events (chat/*, filetransfer/*, ...)
client.events.on('message', (event, payload) => console.log(event, payload))

// point at a server: resumes a saved session (autologin) or logs in with ?token= in the URL
client.setUrl('https://neko.example.com/')
// or log in by hand
await client.login('name', 'password')
client.connect()

// the video, with the input overlay on top
client.mount(document.getElementById('video')!)
```

## `NekoClient`

### Options

| option        | default  |                                                                                     |
| ------------- | -------- | ----------------------------------------------------------------------------------- |
| `autologin`   | `false`  | remember the session token in `localStorage` (`neko_session`) and resume it         |
| `autoconnect` | `false`  | connect as soon as `setUrl()` finds a valid session                                 |
| `autoplay`    | `false`  | start playback when the stream becomes playable (read each time, so a getter works) |
| `inputMode`   | `'auto'` | `'touch'` or `'mouse'` instead of detecting it (`isTouchDevice`)                    |
| `transport`   | WebRTC   | a `StreamTransport` (or a factory for one), see [Transports](#transports)           |

### Authentication

| method                        |                                                                                                          |
| ----------------------------- | -------------------------------------------------------------------------------------------------------- |
| `setUrl(url = location.href)` | the server (defaults to the page itself); `?token=` logs in with that token; with `autoconnect` connects |
| `authenticate(token?)`        | validate a token (or the saved one with `autologin`) with `GET /api/whoami`                              |
| `login(username, password)`   | `POST /api/login`; keeps the token when the server runs without cookies                                  |
| `logout()`                    | `POST /api/logout`, forgets the token                                                                    |

### Connection

`connect()` opens the websocket (`/api/ws`) and, once the server has introduced the session
(`system/init`), starts the transport, which brings up the stream (WebRTC: asks for an offer and
answers it; the server's ICE candidates are buffered until the offer is applied). The connection
is kept alive by itself: a socket that goes silent is replaced and the transport restarted on the
new one, a peer the server drops is requested again, and after repeated failures the client backs
off and finally gives up (`connection.closed` with an error). `disconnect()` closes everything;
the session stays valid.

`connection.status` follows both: `'connecting'` while the socket is being (re)opened or the
transport has no stream yet, `'connected'` once the transport reports `'connected'` (or
`'unavailable'`: the server will not stream to this session, the room is shown without video).

| event / state                                        |                                                 |
| ---------------------------------------------------- | ----------------------------------------------- |
| `state.connection.status`                            | `'disconnected'`, `'connecting'`, `'connected'` |
| `events: 'connection.status'`                        | the status changed                              |
| `events: 'connection.closed'`                        | closed; with an `Error` when not asked for      |
| `connected`, `session`, `isAdmin`, `implicitControl` | computed from the state                         |
| `transport`                                          | the `StreamTransport` in use                    |

### State

`client.state` is a deep-observable object (`Store`); read it anywhere, subscribe for changes:

```ts
const off = client.store.subscribe(() => {}) // after any change, once per tick
client.store.watch(
  () => client.state.control.host_id,
  (hostId, before) => {},
) // when a value changes
client.store.version // increases on every change; what React's useSyncExternalStore compares
```

| field                                      |                                                                                    |
| ------------------------------------------ | ---------------------------------------------------------------------------------- |
| `authenticated`                            | a valid session is known                                                           |
| `connection.{url, token, status}`          |                                                                                    |
| `video.{playable, playing, volume, muted}` | as the transport reports it; plus `mutedByAutoplay` when the browser refused sound |
| `control.host_id`                          | who has control (`controlling` compares it with `session_id`)                      |
| `control.locked`                           | local lock: keep control, send no input                                            |
| `control.clipboard`                        | the remote clipboard, as the server reports it                                     |
| `control.{scroll, keyboard, touch}`        | input settings and whether the server takes native touch events                    |
| `screen.{size, configurations}`            | the current and the available screen sizes                                         |
| `session_id`, `sessions`                   | this session and all sessions (`Session`: `id`, `profile`, `state`)                |
| `settings`                                 | room settings (`Settings`: locks, implicit hosting, `plugins`, ...)                |
| `mobile_keyboard_open`                     | the on-screen keyboard is up (touch devices)                                       |

### Events

`client.events` is a typed `Emitter` (`on`, `off`, `once`; `on` returns the unsubscribe):

| event                    | arguments                                                              |
| ------------------------ | ---------------------------------------------------------------------- |
| `session.created`        | `id`                                                                   |
| `session.updated`        | `id` (profile or state changed)                                        |
| `session.deleted`        | `id`                                                                   |
| `room.control.host`      | `hasHost, hostId, by`                                                  |
| `room.control.request`   | `id` of the member asking for control (hosts only)                     |
| `room.screen.updated`    | `width, height, rate, by`                                              |
| `room.settings.updated`  | `settings, by`                                                         |
| `room.clipboard.updated` | `text`                                                                 |
| `room.broadcast.status`  | `active, url`                                                          |
| `receive.broadcast`      | `sender, subject, body` (`sendBroadcast` from another client)          |
| `receive.unicast`        | `sender, subject, body`                                                |
| `upload.drop.progress`   | `{ loaded, total }` of a file dropped on the video                     |
| `upload.drop.finished`   | `error?`                                                               |
| `overlay.click`          | a click on the video (the GUI hints at the control button with it)     |
| `message`                | `event, payload`: everything else, i.e. plugin events (`chat/message`) |

### Video and audio

The stream is the transport's: `mount(el)` lets it create its media element inside `el` and puts
the input overlay over it; the media calls below go through it.

| method / field                       |                                                                                                          |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------- |
| `mount(el)` / `unmount()`            | put the stream and the input overlay into `el` (they fill it); `canvasSize` is their size                |
| `transport.element`                  | the transport's media element (a `<video>` for WebRTC), once mounted                                     |
| `play()`, `pause()`                  | when the browser refuses sound, `play()` starts muted (`mutedByAutoplay`) and unmutes on the first click |
| `mute()`, `unmute()`, `setVolume(v)` | `v` in 0..1                                                                                              |
| `shareMedia(stream)`                 | send local media (microphone) on a transport that can; returns the function that stops sharing it        |

### Control and input

The overlay sends mouse, wheel, keyboard and touch input through `client.input`, an `InputChannel`:
the transport's own channel when it has one (WebRTC: the data channel, `DataChannelInput`, binary
frames), otherwise `WebSocketInput`, which sends the same input as the server's `control/*`
websocket events. Keyboard handling is Apache Guacamole's keyboard (vendored in `keyboard/`), so
keysyms match the server's layouts.

```ts
interface InputChannel {
  move(x: number, y: number): void // remote screen coordinates
  scroll(deltaX: number, deltaY: number, controlKey: boolean): void
  button(code: number, down: boolean): void // X11 button code
  key(keysym: number, down: boolean): void // X11 keysym
  touch(phase: 'begin' | 'update' | 'end', id: number, x: number, y: number, pressure: number): void
}
```

| method                                           |                                                                                            |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------ |
| `request()`, `release()`                         | ask for / give up control                                                                  |
| `lock()`, `unlock()`                             | local lock (`control.locked`)                                                              |
| `setScrollInverse(v)`, `setScrollSensitivity(n)` | wheel direction and v3 scroll steps (-5..5)                                                |
| `setKeyboard(layout, variant?)`                  | the remote keyboard layout, sent while in control                                          |
| `setScreenSize(width, height, rate)`             | admins: change the screen                                                                  |
| `paste(text)`                                    | type text remotely through the server's clipboard                                          |
| `preparePaste(text)`                             | before a paste keystroke: the side that copied most recently wins (see the source comment) |
| `uploadDrop({ x, y, files })`                    | drop files into the remote desktop at a position (the overlay does this on drop)           |
| `mobileKeyboardToggle()`                         | open / close the on-screen keyboard on touch devices                                       |
| `input.move(x, y)`, `input.key(keysym, down)`, … | input straight to the remote desktop (the `InputChannel` above)                            |

### Messages

| method                                 |                                                            |
| -------------------------------------- | ---------------------------------------------------------- |
| `send(event, payload?)`                | a websocket message, e.g. `send('chat/message', { text })` |
| `sendBroadcast(subject, body)`         | to every other client (`receive.broadcast` there)          |
| `sendUnicast(receiver, subject, body)` | to one session                                             |

### REST

`client.api` (`NekoApi`) carries the server URL and the token:

```ts
const sessions = await client.api.req<Session[]>('GET', '/sessions')
await client.api.req('POST', `/members/${encodeURIComponent(id)}`, { can_login: false })
const blob = await client.api.blob('/room/screen/shot.jpg') // admins
await client.api.upload('/filetransfer', formData, ({ loaded, total }) => {})
```

Failures throw `ApiError` with `status` and the server's `message`. Session ids can come from
login names on some member providers, so they go through `encodeURIComponent` in paths.

## Transports

`NekoClient` owns the session (auth, the websocket, the room state) and leaves the stream to a
`StreamTransport` (`transport.ts`), the way Guacamole's client runs over a tunnel it does not care
about. WebRTC is the one implemented (`WebRTCTransport`, `transport/webrtc.ts`); others
(WebCodecs over a websocket, WebTransport, MSE, HLS) plug in the same way:

```ts
interface StreamTransport {
  readonly kind: 'webrtc' | 'webcodecs-ws' | 'ws-mse' | 'hls'
  readonly element: HTMLElement | null // its media element, created by attach()
  readonly input?: InputChannel // its own input path, if any (the data channel)

  connect(session: SessionInfo): Promise<void> // start streaming; again after close() on a reconnect
  close(): void // drop the stream, timers and counters; the last picture stays on the element

  attach(container: HTMLElement): () => void // create the media element inside; returns what removes it

  setPlaying(on: boolean): Promise<void>
  setVolume(volume: number): void
  setMuted(on: boolean): void
  shareMedia?(stream: MediaStream): () => void // local media (microphone); returns what stops sharing it

  on(event: 'state' | 'stats' | 'error' | 'cursor.position' | 'cursor.image', cb): () => void
}
```

`SessionInfo` is what the client hands over on `system/init`: the server `url` and `token`, the
`init` payload, and `send(event, payload)` / `on(event, cb)` for the transport's own messages over
the main socket (`signal/*` go to the transport, the client does not handle them).

A transport reports `'state'` (`{ status, size, video }`: its `status` — `'connecting'`,
`'connected'`, `'unavailable'` when the server will not stream, `'disconnected'` — the stream's
size and the media element's `playable`/`playing`/`volume`/`muted`/`mutedByAutoplay`, mirrored
into `state.video`), `'error'` (the stream is lost for good; the client closes the connection
with it), the host's `'cursor.position'`/`'cursor.image'` when its server feeds them back, and
`'stats'` (reserved). A transport does not start playback by itself: the client calls
`setPlaying(true)` when `playable` turns true and `autoplay` is on, and a transport keeps playing
across its own reconnects. A second transport needs: its own connection from the `SessionInfo`, an
element it draws into, play/volume/mute, and either an `input` channel or nothing (the client then
sends input over the websocket). Pass it as `new NekoClient({ transport })`; the GUI checks
`WebRTCTransport.supported()` before rendering.

## `Store` and `Emitter`

`new Store(initial)` wraps a plain object in proxies: writes anywhere in the tree mark it dirty, and
one microtask later `watch` callbacks run and `subscribe` listeners are called. Arrays and objects
read from the store are proxies too, so in-place mutation (`push`, `splice`, assignment) is seen.
`Emitter<Events>` is a small typed event emitter. Both are exported for GUIs that want the same
pattern for their own state.

## Types

The wire types mirror `server/pkg/types`: `Session`, `MemberProfile`, `SessionState`, `Settings`,
`ScreenSize`, `InitPayload`, `LoginResponse`, `MemberData`, `CursorImage`, the client `State` and
the `NekoEvents` map; the transport seam is `StreamTransport`, `SessionInfo`, `InputChannel`,
`TransportKind`, `TransportStatus`, `TransportState` and `TransportEvents`.

## Development

```sh
npm run build   # tsc -> dist/ (ESM + .d.ts), plus the vendored keyboard library
npm test        # the unit tests in src/**/*.test.ts (vitest, run from ../ where it is installed): the connection
                # state machine (handshake, reconnects, timeouts, events, clipboard, auth), the transport seam,
                # the WebRTC transport, the input channels and the store, against the fake browser in
                # src/test/browser.ts (sockets, peers, elements, fake timers)
npm run check   # build, the tests, then node dist.check.mjs against the built package: exports, Store,
                # Emitter, setUrl, .d.ts specifiers
```

The tests and dist.check.mjs run under Node with a few browser globals stubbed, so the connection state
machine and the store are tested without a browser. The GUI in `../src` consumes this package from source
through the `@m1k1o/neko` alias in `../tsconfig.json` and `../vite.config.ts`; consumers of the npm
package get `dist/`.
