import { useState } from 'react'
import { useNeko } from '@/state/hooks'
import { actions } from '@/state/actions'
import { isLocked, type LockResource } from '@/state/client'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { Logo } from '@/components/Logo'
import { t } from '@/i18n'
import './header.scss'

export function Header() {
  const { client, app } = useNeko()
  const admin = client.isAdmin
  const [read, setRead] = useState(app.texts)

  const lock = (r: LockResource, icon: string) => {
    const locked = isLocked(r)
    const tip = admin
      ? t(`locks.${r}.${locked ? 'unlock' : 'lock'}`)
      : t(`locks.${r}.${locked ? 'locked' : 'unlocked'}`)
    return (
      <li>
        <i
          className={`fas ${icon}${admin ? '' : ' disabled'}${locked ? ' locked' : ''}`}
          {...a11y(tip)}
          aria-disabled={!admin}
          onClick={() => admin && actions.toggleLock(r)}
        />
      </li>
    )
  }

  return (
    <div className="header">
      <a
        href="https://github.com/m1k1o/neko"
        title="Github repository"
        target="_blank"
        rel="noreferrer"
        className="neko"
      >
        <Logo />
      </a>
      <ul className="menu">
        {lock('control', 'fa-mouse')}
        {lock('login', isLocked('login') ? 'fa-lock' : 'fa-lock-open')}
        {app.files?.enabled && lock('file_transfer', 'fa-file')}
        <li>
          {!app.side && read !== app.texts && <span className="badge">&bull;</span>}
          <i
            className="fas fa-bars toggle"
            {...a11y('Toggle side panel')}
            aria-expanded={app.side}
            onClick={() => {
              app.side = !app.side
              remember('side', app.side)
              setRead(app.texts)
            }}
          />
        </li>
      </ul>
    </div>
  )
}
