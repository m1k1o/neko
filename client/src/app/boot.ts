// start-up, in the order the pieces need: the event wiring that turns server events into event
// lines and toasts, the plugins (strings, event dispatch, their own wiring), the settings (also
// from the URL), then the connection: resume a saved session (autologin) and connect
import '@/state/events'
import { initPlugins } from '@/plugins'
import { initSettings } from '@/state/settings'
import { client } from '@/state/client'

initPlugins()
initSettings()
client.setUrl(location.href)
