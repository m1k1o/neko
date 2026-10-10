import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@/design/index.scss'
import { ready } from './app/boot'
import { App } from './app/App'
import { ErrorBoundary } from './app/ErrorBoundary'

// the first render waits for the strings of the active language (app/boot.ts)
ready.then(() =>
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ErrorBoundary>
        <App />
      </ErrorBoundary>
    </StrictMode>,
  ),
)
