// start-up, in the order the pieces need: the event wiring that turns server events into event
// lines and toasts, the strings of the active language (the plugins' namespaces included), the
// plugins (slots, event dispatch, their own wiring), the settings (also from the URL), then the
// connection: resume a saved session (autologin) and connect
import '@/state/events'
import { initI18n } from '@/i18n'
import { plugins, initPlugins } from '@/plugins'
import { initSettings } from '@/state/settings'
import { client } from '@/state/client'

export const ready = initI18n(plugins.map((p) => p.ns))
initPlugins()
initSettings()
// the connection, like the first render (main.tsx), waits for the strings: event lines and toasts use them
ready.then(() => client.setUrl(location.href))
