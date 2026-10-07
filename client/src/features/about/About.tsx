import { useEffect } from 'react'
import { useNeko } from '@/state/hooks'
import { closeOn } from '@/components/a11y'
import { Logo } from '@/components/Logo'
import { t } from '@/i18n'
import { version } from '../../../package.json'
import './about.scss'

export function About() {
  const { app } = useNeko()
  useEffect(() => closeOn(() => (app.about = false)), [app])
  return (
    <div className="about" onClick={(e) => e.target === e.currentTarget && (app.about = false)}>
      <div className="window" role="dialog" aria-label="About n.eko">
        <div className="about-content">
          <div className="logo">
            <Logo />
          </div>
          <p>A self hosted virtual browser that runs in docker and uses WebRTC.</p>
          <p className="links">
            <a href="https://github.com/m1k1o/neko" target="_blank" rel="noopener noreferrer">
              <i className="fab fa-github" /> m1k1o/neko
            </a>
            <a href="https://neko.m1k1o.net/" target="_blank" rel="noopener noreferrer">
              <i className="fas fa-book" /> Documentation
            </a>
          </p>
          <p className="version">client {version}</p>
          <button onClick={() => (app.about = false)}>{t('connection.button_confirm')}</button>
        </div>
      </div>
    </div>
  )
}
