import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { initI18n } from './i18n'
import { plugins, initPlugins } from './plugins'
import { createNekoApp } from './state/neko'
import { NekoProvider } from './state/provider'
import { App } from './app/App'
import { ErrorBoundary } from './app/ErrorBoundary'

// start-up, in the order the pieces need: the strings of the active language (the plugins'
// namespaces included), the instance (the client with its event wiring that turns server events
// into event lines and toasts, and the settings, also from the URL), the plugins (slots, event
// dispatch, their own wiring), then, once the strings are there (event lines, toasts and the first
// render use them), the connection: resume a saved session (autologin) and connect
const ready = initI18n(plugins.map((p) => p.ns))
const neko = createNekoApp()
initPlugins(neko)
ready.then(() => {
  neko.client.setUrl(location.href)
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ErrorBoundary>
        <NekoProvider neko={neko}>
          <App />
        </NekoProvider>
      </ErrorBoundary>
    </StrictMode>,
  )
})
