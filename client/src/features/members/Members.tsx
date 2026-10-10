import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectSession } from '@m1k1o/neko'
import { cn } from '@/lib/utils'
import { openMenu } from '@/state/actions'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { Avatar } from '@/components/Avatar'

// keyboard equivalent of right-click: the Menu key or Shift+F10 on a focused member
const menuKey = (e: React.KeyboardEvent, id: string) => {
  if (e.key !== 'ContextMenu' && !(e.shiftKey && e.key === 'F10')) return
  e.preventDefault()
  e.stopPropagation()
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  if (id !== client.state.session_id) app.setState({ menu: { x: r.left, y: r.bottom, id } })
}

const member = 'relative mt-2.5 mr-[5px] ml-[5px] block h-[50px] w-[50px]'
// the badges on an avatar: you (a user icon), an admin (a shield), the host (a crown)
const self =
  'fas fa-user-circle absolute mt-[-2px] ml-10 h-[15px] w-[15px] rounded-full bg-background-floating text-center text-[20px] leading-[15px] text-style-primary'
const admin = 'fas fa-shield-alt absolute mt-[-2px] ml-11 block h-3.5 w-3.5 text-center text-[14px] text-style-primary'
const host =
  'fas fa-crown absolute mt-[42px] ml-[-18px] block h-5 w-5 rounded-full bg-style-primary text-center text-[10px] leading-5 text-background-floating'
// the others start after a vertical rule
const others =
  'ml-5 before:absolute before:mt-[13px] before:ml-[-9px] before:h-[45px] before:w-0.5 before:bg-background-secondary before:content-[" "]'

export function Members() {
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
                <i className={self} />
                <Avatar seed={me.profile.name} avatar={me.profile.avatar} size={50} />
                {me.id === hostId && <i className={host} />}
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
                  onContextMenu={(e) => openMenu(e, m.id)}
                  onKeyDown={(e) => menuKey(e, m.id)}
                >
                  {m.profile.is_admin && <i className={admin} />}
                  <Avatar seed={m.profile.name} avatar={m.profile.avatar} size={50} />
                  {m.id === hostId && <i className={host} />}
                </div>
              </li>
            ))}
        </ul>
      </div>
    </div>
  )
}
