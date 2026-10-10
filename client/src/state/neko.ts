// One instance of the GUI: the core client with its input overlay, the GUI store and the bus, wired
// (event lines, toasts, the settings, also from the URL). The application makes one (main.tsx),
// adds the plugins (initPlugins), points it at the server (client.setUrl) and renders under
// <NekoProvider>; two instances share nothing.
import { NekoClient, Overlay, type NekoClientOptions } from '@m1k1o/neko'
import { createAppStore, type NekoApp } from './app'
import { createBus } from './bus'
import { wireEvents } from './events'
import { initSettings } from './settings'

export type NekoAppOptions = Pick<NekoClientOptions, 'transport'>

export function createNekoApp(options: NekoAppOptions = {}): NekoApp {
  const app = createAppStore()
  const client = new NekoClient({
    autologin: true,
    autoconnect: true,
    // read when a track arrives, so the setting applies without a reload
    get autoplay() {
      return app.getState().settings.autoplay
    },
    transport: options.transport,
  })
  const neko: NekoApp = { client, overlay: new Overlay(client), app, bus: createBus() }
  wireEvents(neko)
  initSettings(neko)
  return neko
}
