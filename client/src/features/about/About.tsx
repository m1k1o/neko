import { useEffect } from 'react'
import { app } from '@/state/app'
import { closeOn } from '@/components/a11y'
import { Logo } from '@/components/Logo'
import { t } from '@/i18n'
import './about.scss'

const close = () => app.setState({ about: false })

export function About() {
  useEffect(() => closeOn(close), [])
  return (
    <div className="about" onClick={(e) => e.target === e.currentTarget && close()}>
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
          <p className="version">client {__APP_VERSION__}</p>
          <button onClick={close}>{t('connection.button_confirm')}</button>
        </div>
      </div>
    </div>
  )
}
