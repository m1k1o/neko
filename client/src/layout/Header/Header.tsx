import { useState } from 'react'
import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { app } from '@/state/app'
import { client, isLocked, type LockResource } from '@/state/client'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { LockButton } from '@/components/LockButton'
import { useSlot } from '@/plugins'
import { EachHook } from '@/components/EachHook'
import logo from '@/assets/images/logo.svg'

const zero = () => 0

export function Header() {
  const { t } = useTranslation()
  const admin = useStore(client.store, selectIsAdmin)
  const settings = useStore(client.store, (s) => s.settings)
  const side = useStore(app, (s) => s.side)
  const items = useSlot('header.item')
  const tabs = useSlot('side.tab')

  const lock = (r: LockResource, icon: string) => {
    const locked = isLocked(r, settings)
    const tip = admin
      ? t(`locks.${r}.${locked ? 'unlock' : 'lock'}`)
      : t(`locks.${r}.${locked ? 'locked' : 'unlocked'}`)
    return (
      <LockButton id={r} icon={icon} locked={locked} admin={admin} tip={tip} onToggle={() => actions.toggleLock(r)} />
    )
  }

  return (
    <div className="flex flex-1 flex-row items-center">
      <a
        href="https://github.com/m1k1o/neko"
        title="Github repository"
        target="_blank"
        rel="noreferrer"
        className="ml-5 flex w-[150px] flex-1 items-center justify-start"
      >
        <img src={logo} alt="n.eko" className="mr-2.5 block h-7.5" />
        <span className="text-[30px] leading-7.5">
          <b className="font-black">n</b>.eko
        </span>
      </a>
      <ul className="mr-2.5 whitespace-nowrap">
        {lock('control', 'fa-mouse')}
        {lock('login', isLocked('login', settings) ? 'fa-lock' : 'fa-lock-open')}
        {items.map(({ id, component: Item }) => (
          <Item key={id} />
        ))}
        {/* what the side panel has to show, for the badge on its toggle */}
        <EachHook items={tabs} use={(tab) => tab.useBadge ?? zero}>
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
    <li className="mr-2.5 inline-block">
      {!side && read !== unread && (
        <span className="pointer-events-none absolute h-5 w-5 animate-badge rounded-full bg-[red] text-center text-[1.25em] leading-5 font-bold">
          &bull;
        </span>
      )}
      <i
        className="fas fa-bars block h-7.5 w-7.5 cursor-pointer rounded-[3px] bg-background-primary text-center leading-8"
        data-testid="side-toggle"
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
