# @m1k1o/neko

Framework-free client library for the [neko](https://github.com/m1k1o/neko) v3 API: REST, the
`/api/ws` event protocol, WebRTC with the binary input channel, and the input overlay (mouse,
wheel, keyboard, touch, file drop, the host's cursor). It is what the React GUI in `../src` is
built on, and it can drive a neko server from any framework or from a plain page.

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

| option        | default  |                                                                             |
| ------------- | -------- | --------------------------------------------------------------------------- |
| `autologin`   | `false`  | remember the session token in `localStorage` (`neko_session`) and resume it |
| `autoconnect` | `false`  | connect as soon as `setUrl()` finds a valid session                         |
| `autoplay`    | `false`  | start playback when the track arrives (read each time, so a getter works)   |
| `inputMode`   | `'auto'` | `'touch'` or `'mouse'` instead of detecting it (`isTouchDevice`)            |

### Authentication

| method                        |                                                                                                          |
| ----------------------------- | -------------------------------------------------------------------------------------------------------- |
| `setUrl(url = location.href)` | the server (defaults to the page itself); `?token=` logs in with that token; with `autoconnect` connects |
| `authenticate(token?)`        | validate a token (or the saved one with `autologin`) with `GET /api/whoami`                              |
| `login(username, password)`   | `POST /api/login`; keeps the token when the server runs without cookies                                  |
| `logout()`                    | `POST /api/logout`, forgets the token                                                                    |

### Connection

`connect()` opens the websocket (`/api/ws`), asks for a WebRTC offer and answers it; the server's
ICE candidates are buffered until the offer is applied. The connection is kept alive by itself: a
socket that goes silent is replaced, a peer the server drops is requested again, and after
repeated failures the client backs off and finally gives up (`connection.closed` with an error).
`disconnect()` closes everything; the session stays valid.

| event / state                                        |                                                 |
| ---------------------------------------------------- | ----------------------------------------------- |
| `state.connection.status`                            | `'disconnected'`, `'connecting'`, `'connected'` |
| `events: 'connection.status'`                        | the status changed                              |
| `events: 'connection.closed'`                        | closed; with an `Error` when not asked for      |
| `connected`, `session`, `isAdmin`, `implicitControl` | computed from the state                         |

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

| field                                      |                                                                                  |
| ------------------------------------------ | -------------------------------------------------------------------------------- |
| `authenticated`                            | a valid session is known                                                         |
| `connection.{url, token, status}`          |                                                                                  |
| `video.{playable, playing, volume, muted}` | plus `mutedByAutoplay` when the browser refused sound and playback started muted |
| `control.host_id`                          | who has control (`controlling` compares it with `session_id`)                    |
| `control.locked`                           | local lock: keep control, send no input                                          |
| `control.clipboard`                        | the remote clipboard, as the server reports it                                   |
| `control.{scroll, keyboard, touch}`        | input settings and whether the server takes native touch events                  |
| `screen.{size, configurations}`            | the current and the available screen sizes                                       |
| `session_id`, `sessions`                   | this session and all sessions (`Session`: `id`, `profile`, `state`)              |
| `settings`                                 | room settings (`Settings`: locks, implicit hosting, `plugins`, ...)              |
| `mobile_keyboard_open`                     | the on-screen keyboard is up (touch devices)                                     |

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

| method / field                       |                                                                                                          |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------- |
| `mount(el)` / `unmount()`            | put the `<video>` and the input overlay into `el` (they fill it); `canvasSize` is their size             |
| `video`                              | the `HTMLVideoElement`, once mounted                                                                     |
| `play()`, `pause()`                  | when the browser refuses sound, `play()` starts muted (`mutedByAutoplay`) and unmutes on the first click |
| `mute()`, `unmute()`, `setVolume(v)` | `v` in 0..1                                                                                              |
| `addTrack(track, ...streams)`        | send a local track (microphone); returns the `RTCRtpSender` for `removeTrack(sender)`                    |

### Control and input

The overlay sends mouse, wheel, keyboard and touch input over the data channel while this session
has control. Keyboard handling is Apache Guacamole's keyboard (vendored in `keyboard/`), so
keysyms match the server's layouts.

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
| `sendData(op, ...fields)`                        | raw input message on the data channel; opcodes in `OP`                                     |

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

## `Store` and `Emitter`

`new Store(initial)` wraps a plain object in proxies: writes anywhere in the tree mark it dirty, and
one microtask later `watch` callbacks run and `subscribe` listeners are called. Arrays and objects
read from the store are proxies too, so in-place mutation (`push`, `splice`, assignment) is seen.
`Emitter<Events>` is a small typed event emitter. Both are exported for GUIs that want the same
pattern for their own state.

## Types

The wire types mirror `server/pkg/types`: `Session`, `MemberProfile`, `SessionState`, `Settings`,
`ScreenSize`, `LoginResponse`, `MemberData`, `CursorImage`, the client `State`, the `NekoEvents`
map and the data channel opcodes `OP`.

## Development

```sh
npm run build   # tsc -> dist/ (ESM + .d.ts), plus the vendored keyboard library
npm test        # the unit tests in src/*.test.ts (vitest, run from ../ where it is installed): the connection
                # state machine (handshake, reconnects, timeouts, events, clipboard, auth) and the store,
                # against fake sockets, peers and fake timers
npm run check   # build, the tests, then node dist.check.mjs against the built package: exports, Store,
                # Emitter, setUrl, .d.ts specifiers
```

The tests and dist.check.mjs run under Node with a few browser globals stubbed, so the connection state
machine and the store are tested without a browser. The GUI in `../src` consumes this package from source
through the `@m1k1o/neko` alias in `../tsconfig.json` and `../vite.config.ts`; consumers of the npm
package get `dist/`.
