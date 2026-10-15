// One instance of the GUI: the core client with its input overlay, the GUI store and the bus, wired
// (event lines, toasts, the settings, also from the URL). The application makes one (main.tsx),
// adds the plugins (initPlugins), points it at the server (client.setUrl) and renders under
// <NekoProvider>; two instances share nothing.
import {
  NekoClient,
  Overlay,
  WebCodecsTransport,
  WebRTCTransport,
  type NekoClientOptions,
  type StreamTransport,
} from '@m1k1o/neko'
import { createAppStore, isTransportMode, type NekoApp, type TransportMode } from './app'
import { createBus } from './bus'
import { wireEvents } from './events'
import { initSettings } from './settings'
import { set } from './storage'

export type NekoAppOptions = Pick<NekoClientOptions, 'transport'>

// WebRTC failures to take in 'auto' before switching to the websocket. One retry rides out a
// transient ICE hiccup; past that the media route is blocked and waiting longer only delays video.
const AUTO_FAIL_FAST_AFTER = 1

// The transport to start with, and what to fall back to, for a mode. The fallback lives in the
// client rather than behind a setting because a stream that cannot connect would otherwise leave
// the user unable to reach the settings to change it.
export function transportsFor(mode: TransportMode): Pick<NekoClientOptions, 'transport' | 'fallbackTransport'> {
  const webrtc = WebRTCTransport.supported()
  const websocket = WebCodecsTransport.supported()

  if (mode === 'webrtc' && webrtc) return { transport: () => new WebRTCTransport() }
  if (mode === 'websocket' && websocket) return { transport: () => new WebCodecsTransport() }

  // 'auto', or a forced mode this browser cannot do: use what it can, WebRTC first
  if (webrtc && websocket) {
    return {
      transport: () => new WebRTCTransport({ failFastAfter: AUTO_FAIL_FAST_AFTER }),
      fallbackTransport: () => new WebCodecsTransport(),
    }
  }
  if (websocket) return { transport: () => new WebCodecsTransport() }
  return { transport: () => new WebRTCTransport() }
}

// ?transport= wins over the stored setting and is remembered, so a user whose stream will not
// connect can still change it without first reaching the settings panel
function resolveMode(stored: TransportMode): TransportMode {
  const param = new URL(location.href).searchParams.get('transport')
  if (isTransportMode(param)) {
    set('transport', param)
    return param
  }
  return stored
}

export function createNekoApp(options: NekoAppOptions = {}): NekoApp {
  const app = createAppStore()
  const mode = resolveMode(app.getState().settings.transport)
  if (mode !== app.getState().settings.transport) {
    app.setState((s) => ({ settings: { ...s.settings, transport: mode } }))
  }
  // an explicit transport (an embedder's) overrides the setting entirely
  const transports: Pick<NekoClientOptions, 'transport' | 'fallbackTransport'> = options.transport
    ? { transport: options.transport }
    : transportsFor(mode)

  const client = new NekoClient({
    autologin: true,
    autoconnect: true,
    // read when a track arrives, so the setting applies without a reload
    get autoplay() {
      return app.getState().settings.autoplay
    },
    ...transports,
  })
  const neko: NekoApp = { client, overlay: new Overlay(client), app, bus: createBus() }
  wireEvents(neko)
  initSettings(neko)
  return neko
}

export type { StreamTransport }
