# neko client

The web client for [neko](https://github.com/m1k1o/neko): React 19 + TypeScript, built with Vite.
It speaks only the v3 API (`/api/ws` + REST), so it works with `NEKO_LEGACY=false`.

```
client/
  core/          @m1k1o/neko: the framework-free client library, its own npm package (core/README.md)
  src/
    app/         the shell: App, ErrorBoundary, boot.ts (start-up order), the #neko grid
    layout/      Header, Side (the tab bar), RoomBar (members, room menu, controls, emotes)
    features/    one folder per feature: video, members, controls, room-menu, emotes, settings, connect, about
    plugins/     the registry (index.ts: the list, the slots, event dispatch), the contract (types.ts), chat/, filetransfer/
    components/  shared pieces: Avatar, Dialog, Toasts, Logo, ContextMenu, LockButton, a11y
    state/       the app store, the NekoClient instance, actions, settings, dialogs, event wiring
    i18n/        i18next: initI18n(), setLang(), the loader of the locale files
    locales/     the strings, one folder per language: common.json, chat.json, files.json
    design/      SCSS tokens (_variables), reset, fonts, global styles
    assets/      images
```

## Layering

Dependencies point one way: `app → layout → features → components → design`, with `state` and
`i18n` below everything but `design`. In practice:

- A unit is a folder with its component(s) and its `.scss` next to them, imported by the component.
  Other folders import it through its `index.ts` with `@/` (`import { Video } from '@/features/video'`);
  only a folder's own files are imported relatively.
- Shared components know nothing about features or plugins. Features use each other only through
  their `index.ts`. Nothing imports `app/`.
- Plugins import neither each other nor `features/`; the app shell and the layout reach plugins only
  through the registry `@/plugins`, and features get plugin contributions passed in (they see only
  the contract in `@/plugins/types`).
- The core is imported only from its entry point, `@m1k1o/neko`, never from a path into `core/`.
- State is zustand: the core's `client.store`, the GUI's `app` store (`state/app.ts`) and each plugin's
  `store.ts`. A component selects what it renders, `useStore(client.store, (s) => s.video.playing)`
  (`useShallow` for several fields; the core's `selectControlling`, `selectIsAdmin`, `selectSession` for
  the computed values), and re-renders only when that value changes. Everything outside React reads
  `store.getState()` and writes `store.setState()` with a new object, never in place.
- Where two layers must talk without importing each other, there is a store or a signal in `state/`:
  the room's event lines ("bob took the controls") are emitted on `state/bus.ts` and shown by the chat
  plugin; `logout` is a bus signal every store with per-session data listens to.

`npm run check` enforces this: `eslint.config.js` holds one `no-restricted-imports` rule set per
layer, and `tools/cycles.mjs` fails on an import cycle (type-only imports excepted).

## Plugins

A plugin is a folder under `src/plugins/` with its components, styles, store and actions, described
by one object (`src/plugins/types.ts`); its strings are a namespace of their own under `src/locales/`:

| field        |                                                                                                                                                                                     |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | the prefix of its server events: `message` events named `<id>/...` reach `onEvent`                                                                                                  |
| `ns`         | the namespace of its strings, `src/locales/<lang>/<ns>.json` in every language, used as `t('<ns>:key')`                                                                             |
| `tab`        | a side-panel tab: `id` (the remembered tab), `icon`, `label` (the key of its name, `chat:tab`), `component`, the hooks `useVisible()` and `useBadge()` (unread count)               |
| `topBar`     | items for the header menu; each renders its own `<li>` or nothing                                                                                                                   |
| `memberMenu` | entries of another member's context menu: the hook `useVisible(member)`, `label(member)`, `onClick(member)`                                                                         |
| `onEvent`    | `(...[event, payload])` for its server events, typed by `PluginEvents` (event name to payload, `src/plugins/types.ts`): `if (event === 'chat/init')` narrows `payload` to its shape |
| `init`       | called once at start-up, before the connection: subscribe to `client.events`, the bus, ...                                                                                          |

`src/plugins/index.ts` lists the plugins of the build in side-panel order, and `initPlugins()`
(called from `app/boot.ts`) wires them in that order: it puts each `tab`, `topBar` item and
`memberMenu` entry into its slot (`registerSlot('side.tab' | 'header.item' | 'member.menu', item)`),
builds the event dispatch (a map from the `id` prefix to the plugin, one lookup per message) and runs
each `init`. The side panel, the header and the app shell read the slots with `useSlot(name)`
(`useVisible` and `useBadge` are hooks over the plugin's store, so each runs in a small component of
its own and re-renders only that).

To add a plugin:

1. Create `src/plugins/<name>/` with an `index.ts` exporting the plugin object, its components with
   their `.scss`, and a `store.ts` made with zustand's `createStore()` (components select from it with
   `useStore(store, (s) => s.field)`, the rest reads `store.getState()` and writes `store.setState()`).
2. Add its strings as `src/locales/<lang>/<ns>.json` in every language (see i18n).
3. Add the import and the entry to `src/plugins/index.ts`.

Removing the entry removes the tab, the header items, the menu entries and the event handling;
nothing else changes (its locale files are then unused).

## i18n

Strings live in `src/locales/<lang>/<ns>.json`: one folder per language, one file per namespace,
`common.json` (the GUI), `chat.json` and `files.json` (the plugins). The runtime is `i18next` with
`react-i18next` (`src/i18n/index.ts`): components call `useTranslation()` and render
`t('side.settings')` or `t('chat:tab')` (`common` is the default namespace, a plugin's is prefixed);
code outside components imports `t` from `@/i18n`. Values interpolate with `{{name}}`.

Only the active language is downloaded: the files are lazy chunks (`import.meta.glob`), and the
loader hands i18next the `<lang>/<ns>` it asks for at start-up (`initI18n`, before the first render)
and on a pick in the room menu (`setLang`, i18next's `changeLanguage`). `en` is the fallback for a
key a language lacks, so it is loaded next to the active language; the other languages never are
(`tools/dist.check.mjs`, run by `npm run build`, checks the main chunk carries none). The language is
kept in `localStorage` (`lang`, the short code) and set by `?lang=`; `<html lang>` gets the BCP 47
tag (`cn` is `zh-CN`). In development, a key no language has is logged (`i18n: missing ...`).

- Adding a key: add it to `en` and to the other fourteen files. `src/i18n/i18n.test.ts` lists the
  keys each language lacks today and fails on a new one.
- Adding a language: a folder with the three files, and its code in `langs` (`src/i18n/index.ts`,
  the picker order).
- Not yet: plural forms (`key_one`, `key_other`); the strings with a count keep one form, as before.

## core

`core/` is `@m1k1o/neko`: `NekoClient` (auth, websocket with reconnect and stale detection, WebRTC,
the binary input channel, control, clipboard, broadcast), the input `Overlay`, the zustand store with
its selectors, `Emitter`, `NekoApi` and the wire types. The GUI consumes it from source through the `@m1k1o/neko` alias in
`tsconfig.json` and `vite.config.ts`; `npm run build:core` builds the package (`core/dist/`). The API
reference is `core/README.md`.

```ts
import { NekoClient } from '@m1k1o/neko'

const client = new NekoClient({ autologin: true, autoconnect: true, autoplay: true })
client.store.subscribe((s) => s.sessions, render) // state: connection, sessions, control, screen, ...
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

| script                |                                                                                                                                                                    |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `npm run build`       | type-check + production build into `dist/`, then `tools/dist.check.mjs` (no language in the main chunk)                                                            |
| `npm run build:core`  | build the `@m1k1o/neko` package into `core/dist/`                                                                                                                  |
| `npm run check`       | types, formatting, lint (hooks, layering), import cycles, the tests, the built core package                                                                        |
| `npm test`            | unit tests (vitest): the core's connection state machine and store, the chat's markdown parser, the i18n runtime and locale files, the plugin registry; no browser |
| `npm run format`      | prettier                                                                                                                                                           |
| `npm run build:emoji` | regenerate emoji data (`public/emoji.json`, the chat's sprite sheet)                                                                                               |
| `npm run test:e2e`    | two-user browser test against a running server, see below                                                                                                          |

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

The typing step expects a launcher on Alt+F2 (KRunner on KDE, the app finder on Xfce). Set
`NEKO_FILE_URL` to a second server using the `file` member provider (members `alice`/`alice` admin,
`bob`/`bob`) to also test ban/unban. To run the steps in a real Chromium-based browser instead of
Playwright's, start it with `--headless --remote-debugging-port=9222 --user-data-dir=<empty dir>
--autoplay-policy=no-user-gesture-required --use-fake-ui-for-media-stream
--use-fake-device-for-media-stream` and set `NEKO_CDP=http://127.0.0.1:9222`.

## Behaviour notes

- **Kick / ban** follow common chat-app semantics, see the comment at `actions.kick` in
  `src/state/actions.ts`: kick removes the session (the user may log in again), ban sets
  `can_login: false` and is only offered when the auth provider stores accounts.
- Settings use the same `localStorage` keys and format as the previous client, so preferences carry
  over. Scroll sensitivity now uses v3 steps (-5..5, key `scroll_sensitivity`); `?scroll=` is mapped.
- Chat markdown is parsed to React elements (no HTML strings); links are limited to http(s)/mailto.
- File downloads are plain links so the browser streams them; the session token is added to the
  URL only when the server runs without cookies.
