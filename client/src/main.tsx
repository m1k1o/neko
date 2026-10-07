import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@/design/index.scss'
import '@/assets/styles/vendor/_emote.scss'
import '@/assets/styles/vendor/_emoji.scss'
import { App, ErrorBoundary } from './app/App'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <App />
    </ErrorBoundary>
  </StrictMode>,
)
