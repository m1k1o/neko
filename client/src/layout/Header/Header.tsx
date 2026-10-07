import { useState } from 'react'
import { useNeko } from '@/state/hooks'
import { actions } from '@/state/actions'
import { isLocked, type LockResource } from '@/state/client'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { Logo } from '@/components/Logo'
import { LockButton } from '@/components/LockButton'
import { t } from '@/i18n'
import { tabs, topBar } from '@/plugins'
import './header.scss'

// what the side panel has to show, for the badge on its toggle
const unread = () => tabs().reduce((n, tab) => n + (tab.badge?.() ?? 0), 0)

export function Header() {
  const { client, app } = useNeko()
  const admin = client.isAdmin
  const [read, setRead] = useState(unread)

  const lock = (r: LockResource, icon: string) => {
    const locked = isLocked(r)
    const tip = admin
      ? t(`locks.${r}.${locked ? 'unlock' : 'lock'}`)
      : t(`locks.${r}.${locked ? 'locked' : 'unlocked'}`)
    return <LockButton icon={icon} locked={locked} admin={admin} tip={tip} onToggle={() => actions.toggleLock(r)} />
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
        {topBar().map(({ id, component: Item }) => (
          <Item key={id} />
        ))}
        <li>
          {!app.side && read !== unread() && <span className="badge">&bull;</span>}
          <i
            className="fas fa-bars toggle"
            {...a11y('Toggle side panel')}
            aria-expanded={app.side}
            onClick={() => {
              app.side = !app.side
              remember('side', app.side)
              setRead(unread())
            }}
          />
        </li>
      </ul>
    </div>
  )
}
