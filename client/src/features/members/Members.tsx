import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectSession } from '@m1k1o/neko'
import { CircleUser, Crown, Shield } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import type { NekoApp } from '@/state/app'
import { useNeko } from '@/state/provider'
import { Avatar } from '@/components/Avatar'

// keyboard equivalent of right-click: the Menu key or Shift+F10 on a focused member
const menuKey = ({ client, app }: NekoApp, e: React.KeyboardEvent, id: string) => {
  if (e.key !== 'ContextMenu' && !(e.shiftKey && e.key === 'F10')) return
  e.preventDefault()
  e.stopPropagation()
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  if (id !== client.state.session_id) app.setState({ menu: { x: r.left, y: r.bottom, id } })
}

const member =
  'relative mt-2.5 mr-[5px] ml-[5px] block h-[50px] w-[50px] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-style-primary'
// the badges on an avatar: you (a user icon), an admin (a shield), the host (a crown)
const self = 'absolute mt-[-2px] ml-10 size-[15px] rounded-full bg-background-floating text-style-primary'
const admin = 'absolute mt-[-2px] ml-11 size-3.5 text-style-primary'
const host = 'absolute mt-[42px] ml-[-18px] size-5 rounded-full bg-style-primary p-[3px] text-background-floating'
// the others start after a vertical rule
const others =
  'ml-5 before:absolute before:mt-[13px] before:ml-[-9px] before:h-[45px] before:w-0.5 before:bg-background-secondary before:content-[" "]'

export function Members() {
  const neko = useNeko()
  const { client } = neko
  const actions = useActions()
  const { me, hostId, sessions, sessionId } = useStore(
    client.store,
    useShallow((s) => ({
      me: selectSession(s),
      hostId: s.control.host_id,
      sessions: s.sessions,
      sessionId: s.session_id,
    })),
  )
  return (
    <div className="flex min-h-[74px] flex-1 overflow-x-scroll overflow-y-hidden pb-3.5 [scrollbar-color:var(--color-background-secondary)_var(--color-background-tertiary)] [scrollbar-width:thin]">
      <div className="mx-auto block px-5">
        <ul className="whitespace-nowrap">
          {me && (
            <li className="inline-block">
              <div className={member} data-testid="member" data-self data-host={me.id === hostId || undefined}>
                <CircleUser className={self} />
                <Avatar seed={me.profile.name} avatar={me.profile.avatar} size={50} />
                {me.id === hostId && <Crown className={host} />}
              </div>
            </li>
          )}
          {Object.values(sessions)
            .filter((m) => m.id !== sessionId && m.state.is_connected)
            .map((m, i) => (
              <li key={m.id} className={cn('inline-block', i === 0 && others)} title={m.profile.name}>
                <div
                  className={member}
                  data-testid="member"
                  data-host={m.id === hostId || undefined}
                  tabIndex={0}
                  aria-label={m.profile.name}
                  aria-haspopup="menu"
                  onContextMenu={(e) => actions.openMenu(e, m.id)}
                  onKeyDown={(e) => menuKey(neko, e, m.id)}
                >
                  {m.profile.is_admin && <Shield className={admin} />}
                  <Avatar seed={m.profile.name} avatar={m.profile.avatar} size={50} />
                  {m.id === hostId && <Crown className={host} />}
                </div>
              </li>
            ))}
        </ul>
      </div>
    </div>
  )
}
