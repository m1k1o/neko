// start-up, in the order the pieces need: the event wiring that turns server events into chat
// lines and toasts, the settings (also from the URL), then the connection: resume a saved
// session (autologin) and connect
import '@/state/events'
import { initSettings } from '@/state/settings'
import { client } from '@/state/client'

initSettings()
client.setUrl(location.href)
