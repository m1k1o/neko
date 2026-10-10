# neko client

The web client for [neko](https://github.com/m1k1o/neko): React 19 + TypeScript, built with Vite.
It speaks only the v3 API (`/api/ws` + REST), so it works with `NEKO_LEGACY=false`.

```
client/
  core/          @m1k1o/neko: the framework-free client library, its own npm package (core/README.md)
  src/
    main.tsx     start-up: the strings, the instance (createNekoApp), the plugins, the connection, the render under <NekoProvider>
    app/         the shell: App, ErrorBoundary, the #neko grid
    layout/      Header, Side (the tab bar), RoomBar (members, room menu, controls, emotes)
    features/    one folder per feature: video, members, controls, room-menu, emotes, settings, connect, about
    plugins/     the registry (index.ts: the list, the slots, event dispatch), the contract (types.ts), chat/, filetransfer/
    components/  shared pieces: IconButton, Avatar, Dialog, EachHook; ui/ holds the shadcn/ui kit (dialog, menus, popover, tabs, button, sonner)
    state/       the instance (neko.ts: createNekoApp, app.ts: the GUI store and NekoApp), the provider and hooks, actions, settings, dialogs, event wiring
    i18n/        i18next: initI18n(), setLang(), the loader of the locale files
    lib/         cn(): class names with Tailwind conflicts resolved (clsx + tailwind-merge)
    locales/     the strings, one folder per language: common.json, chat.json, files.json
    assets/      images
    index.css    Tailwind v4: the design tokens (@theme), the global rules, the layout's breakpoint variants
```

## Layering

Dependencies point one way: `app → layout → features → components → design`, with `state`, `i18n`
and `lib` below everything but `design`. In practice:

- A feature is a folder with its component(s), styled with Tailwind classes in the markup. Other
  folders import a feature by its main file with `@/` (`import { Stage } from '@/features/video/Stage'`);
  only a folder's own files are imported relatively.
- Shared components know nothing about features or plugins. Nothing imports `app/`.
- Plugins import neither each other nor `features/`; the app shell and the layout reach plugins only
  through the registry `@/plugins`, and features get plugin contributions passed in (they see only
  the contract in `@/plugins/types`).
- The core is imported only from its entry point, `@m1k1o/neko`, never from a path into `core/`.
- State is zustand: the core's `client.store`, the GUI's `app` store (`state/app.ts`) and each plugin's
  `store.ts`, all per instance (see Composition). A component gets the instance from the context
  (`useClient()`, `useApp()`, `useNeko()`), selects what it renders,
  `useStore(useClient().store, (s) => s.video.playing)` (`useShallow` for several fields; the core's
  `selectControlling`, `selectIsAdmin`, `selectSession` for the computed values), and re-renders only
  when that value changes. Everything outside React takes the instance as its first argument, reads
  `store.getState()` and writes `store.setState()` with a new object, never in place.
- Where two layers must talk without importing each other, there is a store or a signal in `state/`:
  the room's event lines ("bob took the controls") are emitted on `state/bus.ts` and shown by the chat
  plugin; `logout` is a bus signal every store with per-session data listens to.

`npm run check` enforces this: `eslint.config.js` holds one `no-restricted-imports` rule set per
layer, and `tools/cycles.mjs` fails on an import cycle (type-only imports excepted).

## Composition

There is no global client. `createNekoApp()` (`src/state/neko.ts`) makes one instance of the GUI, a
`NekoApp` (`src/state/app.ts`): the `NekoClient` (WebRTC transport, `autologin`, `autoconnect`,
`autoplay` from the viewer settings), its input `Overlay`, the GUI store `app` and the `bus`, with the
event wiring (server events to event lines and toasts, `state/events.ts`) and the settings (stored
and from the URL, `state/settings.ts`) applied. `main.tsx` composes the page in the order the pieces
need:

```tsx
const ready = initI18n(plugins.map((p) => p.ns)) // the strings of the active language
const neko = createNekoApp() // the client, its wiring, the settings
initPlugins(neko) // the slots, the event dispatch, each plugin's init
ready.then(() => {
  neko.client.setUrl(location.href) // the connection: resume a saved session and connect
  createRoot(root).render(
    <NekoProvider neko={neko}>
      <App />
    </NekoProvider>,
  )
})
```

`<NekoProvider>` (`src/state/provider.tsx`) puts the instance in a React context; a component reads it
with `useNeko()` (the whole `NekoApp`), `useClient()`, `useApp()` (the GUI store) and `useActions()`
(`state/actions.ts`, the room actions bound to the instance), and the rest, `setSetting(neko, ...)`,
`ask(app, ...)`, `api(neko, ...)`, a plugin's `store(neko)`, takes it as the first argument. Two
instances on one page share nothing but the page (the strings, the URL parameters, `localStorage`).

To embed the GUI in another React app, compose the same pieces and render `<App />` under a provider
(`?embed=1` and `?cast=1` on the page's URL still give the video-only layouts, for an iframe or a kiosk):

```tsx
import { initI18n } from '@/i18n'
import { plugins, initPlugins } from '@/plugins'
import { createNekoApp } from '@/state/neko'
import { NekoProvider } from '@/state/provider'
import { App } from '@/app/App'

await initI18n(plugins.map((p) => p.ns))
const neko = createNekoApp()
initPlugins(neko)
neko.client.setUrl('https://neko.example.com/')

export const Room = () => (
  <NekoProvider neko={neko}>
    <App />
  </NekoProvider>
)
```

## Design

Styling is Tailwind v4 (`tailwindcss`, `@tailwindcss/vite`): the classes are in the markup, there is no
stylesheet per component. `src/index.css` is the whole of the CSS:

- **Tokens**: an `@theme` block with the values of the previous SCSS variables under the same names, so
  `$background-primary` is `--color-background-primary` and the utility `bg-background-primary`;
  `$text-normal` is `text-text-normal`, `$style-primary` is `bg-style-primary` / `text-style-primary`,
  `$elevation-high` is `shadow-elevation-high`, `$side-width` / `$menu-height` / `$controls-height` are
  `w-side` / `h-menu` / `h-controls`, `$text-size` is `text-ui`. The font is Tailwind's system stack
  (`--font-sans`). Sizes are written in px where the legacy layout had px (`p-[5px]`, `h-7.5`).
- **Global rules** the preflight does not have: the page's size and overflow, 14px text with line-height 1
  on `body` (so `rem`, the base of Tailwind's spacing scale, stays 16px), no native video controls.
- **Breakpoints** as variants: `tablet:` (up to 1024px: the page scrolls and the side panel goes under the
  video, with Tailwind's `portrait:` / `landscape:` for the split), `phone:` (up to 768px: no room bar).
- The animations (`animate-shake`, `animate-loader`, `animate-badge`) and the `slider` utility (the range
  inputs).

`cn()` (`src/lib/utils.ts`, `clsx` + `tailwind-merge`) joins class names and lets the last Tailwind class
win, so a caller's `className` overrides a component's; `class-variance-authority` gives the variants.

The component kit under `src/components/ui/` is shadcn/ui on Radix, restyled with the tokens: `dialog.tsx`
(`@radix-ui/react-dialog`, the overlay and the content composed by the caller), `dropdown-menu.tsx` (a
menu from a button: the resolution list, the emote picker), `popover.tsx` (the clipboard textarea, the
emoji picker), `tabs.tsx` (the side panel), `context-menu.tsx` (the member menu: a menu opened at a
point, on `@radix-ui/react-menu`, since it is opened from the member list, a chat author and Shift+F10),
`button.tsx` (the text buttons: `primary`, `outline`, `confirm`, `cancel`) and `sonner.tsx` (the toasts,
`toast()` in `src/state/dialogs.ts`). They bring focus trapping, roving focus, Escape, outside click and
ARIA. `components.json` configures `npx shadcn add <component>` (the radix base); a component added
that way comes with shadcn's theme classes and is rewritten with the tokens, like the ones here.

Every icon control is `IconButton` (`src/components/IconButton.tsx`): a `<button type="button">` named
by its `label` (the accessible name and the tooltip), with the icon as its child and the variants `plain`
(the room bar, the chat, the files), `header` (a 30px box) and `video` (the translucent box over the
video). Icons are `lucide-react` components, tree-shaken: `import { Mouse } from 'lucide-react'` and
`<Mouse className="size-4" />` (`size-3.5` next to 14px text, `size-6` in the room bar); a plugin tab's
`icon` is one. A brand icon lucide lacks (GitHub, in About) is an inline `<svg>`. Emoji are text: the
character of each name from `public/emoji.json` (`src/plugins/chat/Emoji.tsx`).

To add a component: Tailwind classes and the tokens in the markup, repeated class strings as constants
next to the component, `Button` / `IconButton` for buttons, the kit for a dialog, a menu or a popover. To
add an icon: import it from `lucide-react`.

## Plugins

A plugin is a folder under `src/plugins/` with its components, styles, store and actions, described
by one object (`src/plugins/types.ts`); its strings are a namespace of their own under `src/locales/`:

| field        |                                                                                                                                                                                           |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | the prefix of its server events: `message` events named `<id>/...` reach `onEvent`                                                                                                        |
| `ns`         | the namespace of its strings, `src/locales/<lang>/<ns>.json` in every language, used as `t('<ns>:key')`                                                                                   |
| `tab`        | a side-panel tab: `id` (the remembered tab), `icon`, `label` (the key of its name, `chat:tab`), `component`, the hooks `useVisible()` and `useBadge()` (unread count)                     |
| `topBar`     | items for the header menu; each renders its own `<li>` or nothing                                                                                                                         |
| `memberMenu` | entries of another member's context menu: the hook `useVisible(member)`, `label(member)`, `onClick(member)`                                                                               |
| `onEvent`    | `(neko, ...[event, payload])` for its server events, typed by `PluginEvents` (event name to payload, `src/plugins/types.ts`): `if (event === 'chat/init')` narrows `payload` to its shape |
| `init`       | `(neko)`, called once per instance at start-up, before the connection: subscribe to `neko.client.events`, `neko.bus`, ...                                                                 |

Every callback gets the instance (`neko: NekoApp`, see Composition) as its first argument, the member
menu's `label(neko, member)` and `onClick(neko, member)` included; the hooks (`useVisible`, `useBadge`)
and the components read it from the context (`useNeko()`, `useClient()`). A plugin's store is one per
instance: `export const store = scoped(() => createStore(...))` (`scoped`, `src/state/app.ts`) and
`store(neko)` is the instance's.

`src/plugins/index.ts` lists the plugins of the build in side-panel order, and `initPlugins(neko)`
(called from `main.tsx`) wires them in that order: it puts each `tab`, `topBar` item and
`memberMenu` entry into its slot (`registerSlot('side.tab' | 'header.item' | 'member.menu', item)`,
once: the slots are the build's, not the instance's), builds the event dispatch (a map from the `id`
prefix to the plugin, one lookup per message) on the instance's `client.events` and runs each
`init(neko)`. The side panel, the header and the app shell read the slots with `useSlot(name)`
(`useVisible` and `useBadge` are hooks over the plugin's store, so each runs in a small component of
its own and re-renders only that).

To add a plugin:

1. Create `src/plugins/<name>/` with an `index.ts` exporting the plugin object, its components (Tailwind
   classes, the tokens of `src/index.css`), and a `store.ts` made with `scoped(() => createStore(...))`
   (components select from it with `useStore(store(useNeko()), (s) => s.field)`, the rest reads
   `store(neko).getState()` and writes `store(neko).setState()`).
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

| script                |                                                                                                                                                                                                                                                      |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `npm run build`       | type-check + production build into `dist/`, then `tools/dist.check.mjs` (no language in the main chunk)                                                                                                                                              |
| `npm run build:core`  | build the `@m1k1o/neko` package into `core/dist/`                                                                                                                                                                                                    |
| `npm run check`       | `check:types`, `check:format`, `check:lint` (hooks, layering), `check:cycles` and `test` in parallel (`npm-run-all2`, one label per script), then `build:core` and `check:dist` (the built core package); each `check:*` script also runs on its own |
| `npm test`            | unit tests (vitest): the core's connection state machine and store, the chat's markdown parser, the i18n runtime and locale files, the plugin registry, an instance and two of them, the provider; no browser                                        |
| `npm run format`      | prettier                                                                                                                                                                                                                                             |
| `npm run build:emoji` | regenerate the emoji data `public/emoji.json` (names, characters, groups, keywords; `tools/emoji.ts`)                                                                                                                                                |
| `npm run test:e2e`    | two-user browser test against a running server, see below                                                                                                                                                                                            |

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
- Short-lived list items (uploads, emote animations) are keyed by `crypto.randomUUID()` (a secure
  context: https or localhost).
- File downloads are plain links so the browser streams them; the session token is added to the
  URL only when the server runs without cookies.
