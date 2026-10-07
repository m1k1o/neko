import { useNeko } from '@/state/hooks'
import { openMenu } from '@/state/actions'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { Avatar } from '@/components/Avatar'
import './members.scss'

// keyboard equivalent of right-click: the Menu key or Shift+F10 on a focused member
const menuKey = (e: React.KeyboardEvent, id: string) => {
  if (e.key !== 'ContextMenu' && !(e.shiftKey && e.key === 'F10')) return
  e.preventDefault()
  e.stopPropagation()
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  if (id !== client.state.session_id) app.state.menu = { x: r.left, y: r.bottom, id }
}

export function Members() {
  const { state } = useNeko()
  const me = client.session
  const host = state.control.host_id
  return (
    <div className="members">
      <div className="members-container">
        <ul className="members-list">
          {me && (
            <li>
              <div className={`member self${me.id === host ? ' host' : ''}`}>
                <Avatar seed={me.profile.name} avatar={me.profile.avatar} size={50} />
              </div>
            </li>
          )}
          {Object.values(state.sessions)
            .filter((m) => m.id !== state.session_id && m.state.is_connected)
            .map((m) => (
              <li key={m.id} title={m.profile.name}>
                <div
                  className={`member${m.id === host ? ' host' : ''}${m.profile.is_admin ? ' admin' : ''}`}
                  tabIndex={0}
                  aria-label={m.profile.name}
                  aria-haspopup="menu"
                  onContextMenu={(e) => openMenu(e, m.id)}
                  onKeyDown={(e) => menuKey(e, m.id)}
                >
                  <Avatar seed={m.profile.name} avatar={m.profile.avatar} size={50} />
                </div>
              </li>
            ))}
        </ul>
      </div>
    </div>
  )
}
