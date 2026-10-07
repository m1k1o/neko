import { useEffect, useRef, useState } from 'react'
import { useNeko } from '@/state/hooks'
import { actions } from '@/state/actions'
import { client } from '@/state/client'
import { ask } from '@/state/dialogs'
import { a11y, closeOn } from '@/components/a11y'
import { Avatar } from '@/components/Avatar'
import { ContextMenu } from '@/components/ContextMenu'
import { t } from '@/i18n'
import type { PluginMemberMenuItem } from '@/plugins/types'

const confirmThen = (title: string, text: string, fn: () => void) => ask(title, text).then((ok) => ok && fn())

// rendered by App outside .room-container, which is hidden at narrow widths where the chat still works
export function MemberMenu({ items }: { items: PluginMemberMenuItem[] }) {
  const { app, state } = useNeko()
  const [bannable, setBannable] = useState(false)
  useEffect(() => closeOn(() => (app.menu = null)), [app])
  // keyboard users land on the first item
  const list = useRef<HTMLUListElement>(null)
  useEffect(() => {
    if (app.menu) list.current?.querySelector<HTMLElement>('[role=menuitem]')?.focus()
  }, [app.menu])

  const id = app.menu?.id
  const m = id ? state.sessions[id] : undefined
  // ban only sticks where the auth provider stores accounts, see actions.ban
  useEffect(() => {
    setBannable(false)
    let current = true // a slow answer for the previous member must not land on this one
    if (id && client.isAdmin) actions.canBan(id).then((v) => current && setBannable(v))
    return () => void (current = false)
  }, [id])
  if (!app.menu || !m) return null

  const n = m.profile.name
  const admin = client.isAdmin
  const implicit = state.settings.implicit_hosting
  const isHost = id === state.control.host_id
  const x = Math.min(app.menu.x, innerWidth - 170)
  const y = Math.min(app.menu.y, innerHeight - 250)

  return (
    <ContextMenu ref={list} role="menu" aria-label={n} style={{ left: x, top: y }}>
      <li className="header">
        <div className="user">
          <Avatar seed={n} avatar={m.profile.avatar} size={25} />
          <strong>{n}</strong>
        </div>
      </li>
      <li className="seperator" />
      <li>
        <span
          {...a11y(t(app.ignored[id!] ? 'context.unignore' : 'context.ignore'), 'menuitem')}
          onClick={() => (app.ignored[id!] = !app.ignored[id!])}
        >
          {t(app.ignored[id!] ? 'context.unignore' : 'context.ignore')}
        </span>
      </li>
      {items
        .filter((item) => !item.visible || item.visible(m))
        .map((item) => (
          <li key={item.id}>
            <span {...a11y(item.label(m), 'menuitem')} onClick={() => item.onClick(m)}>
              {item.label(m)}
            </span>
          </li>
        ))}
      {admin ? (
        <>
          {!implicit && isHost && (
            <>
              <li>
                <span {...a11y(t('context.release'), 'menuitem')} onClick={actions.reset}>
                  {t('context.release')}
                </span>
              </li>
              <li>
                <span {...a11y(t('context.take'), 'menuitem')} onClick={actions.take}>
                  {t('context.take')}
                </span>
              </li>
            </>
          )}
          {!implicit && !isHost && (
            <li>
              <span {...a11y(t('context.give'), 'menuitem')} onClick={() => actions.give(id!)}>
                {t('context.give')}
              </span>
            </li>
          )}
        </>
      ) : (
        client.controlling &&
        !implicit && (
          <li>
            <span {...a11y(t('context.give'), 'menuitem')} onClick={() => actions.give(id!)}>
              {t('context.give')}
            </span>
          </li>
        )
      )}
      {admin && !m.profile.is_admin && (
        <>
          <li className="seperator" />
          <li>
            <span
              style={{ color: '#f04747' }}
              {...a11y(t('context.kick'), 'menuitem')}
              onClick={() =>
                confirmThen(
                  t('context.confirm.kick_title', { name: n }),
                  t('context.confirm.kick_text', { name: n }),
                  () => actions.kick(id!),
                )
              }
            >
              {t('context.kick')}
            </span>
          </li>
          {bannable && (
            <li>
              <span
                style={{ color: '#f04747' }}
                {...a11y(t('context.ban'), 'menuitem')}
                onClick={() =>
                  confirmThen(
                    t('context.confirm.ban_title', { name: n }),
                    t('context.confirm.ban_text', { name: n }),
                    () => actions.ban(id!),
                  )
                }
              >
                {t('context.ban')}
              </span>
            </li>
          )}
        </>
      )}
    </ContextMenu>
  )
}
