import { useState } from 'react'
import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { app } from '@/state/app'
import { client, isLocked, type LockResource } from '@/state/client'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { Logo } from '@/components/Logo'
import { LockButton } from '@/components/LockButton'
import { tabs, topBar } from '@/plugins'
import { EachHook } from '@/components/EachHook'
import './header.scss'

const zero = () => 0

export function Header() {
  const { t } = useTranslation()
  const admin = useStore(client.store, selectIsAdmin)
  const settings = useStore(client.store, (s) => s.settings)
  const side = useStore(app, (s) => s.side)

  const lock = (r: LockResource, icon: string) => {
    const locked = isLocked(r, settings)
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
        {lock('login', isLocked('login', settings) ? 'fa-lock' : 'fa-lock-open')}
        {topBar().map(({ id, component: Item }) => (
          <Item key={id} />
        ))}
        {/* what the side panel has to show, for the badge on its toggle */}
        <EachHook items={tabs()} use={(tab) => tab.useBadge ?? zero}>
          {(counts) => <Toggle side={side} unread={counts.reduce((n, c) => n + c, 0)} />}
        </EachHook>
      </ul>
    </div>
  )
}

// the side panel's toggle, with a badge when the unread count changed while the panel was closed
function Toggle({ side, unread }: { side: boolean; unread: number }) {
  const [read, setRead] = useState(unread)
  return (
    <li>
      {!side && read !== unread && <span className="badge">&bull;</span>}
      <i
        className="fas fa-bars toggle"
        {...a11y('Toggle side panel')}
        aria-expanded={side}
        onClick={() => {
          app.setState({ side: !side })
          remember('side', !side)
          setRead(unread)
        }}
      />
    </li>
  )
}
