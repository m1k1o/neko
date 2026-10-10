import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin, type Session } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { ask } from '@/state/dialogs'
import { a11y, closeOn } from '@/components/a11y'
import { Avatar } from '@/components/Avatar'
import { ContextMenu } from '@/components/ContextMenu'
import type { PluginMemberMenuItem } from '@/plugins/types'

const confirmThen = (title: string, text: string, fn: () => void) => ask(title, text).then((ok) => ok && fn())
const always = () => true

// a plugin's entry, when its hook shows it for this member
function Item({ item, member }: { item: PluginMemberMenuItem; member: Session }) {
  const useVisible = item.useVisible ?? always
  const visible = useVisible(member)
  if (!visible) return null
  return (
    <li>
      <span {...a11y(item.label(member), 'menuitem')} onClick={() => item.onClick(member)}>
        {item.label(member)}
      </span>
    </li>
  )
}

// rendered by App outside .room-container, which is hidden at narrow widths where the chat still works
export function MemberMenu({ items }: { items: PluginMemberMenuItem[] }) {
  const { t } = useTranslation()
  const menu = useStore(app, (s) => s.menu)
  const [bannable, setBannable] = useState(false)
  useEffect(() => closeOn(() => app.setState({ menu: null })), [])
  // keyboard users land on the first item
  const list = useRef<HTMLUListElement>(null)
  useEffect(() => {
    if (menu) list.current?.querySelector<HTMLElement>('[role=menuitem]')?.focus()
  }, [menu])

  const id = menu?.id
  const { m, admin, implicit, hostId, hosting } = useStore(
    client.store,
    useShallow((s) => ({
      m: id ? s.sessions[id] : undefined,
      admin: selectIsAdmin(s),
      implicit: s.settings.implicit_hosting,
      hostId: s.control.host_id,
      hosting: selectControlling(s),
    })),
  )
  const ignored = useStore(app, (s) => !!id && !!s.ignored[id])
  // ban only sticks where the auth provider stores accounts, see actions.ban
  useEffect(() => {
    setBannable(false)
    let current = true // a slow answer for the previous member must not land on this one
    if (id && client.isAdmin) actions.canBan(id).then((v) => current && setBannable(v))
    return () => void (current = false)
  }, [id])
  if (!menu || !m) return null

  const n = m.profile.name
  const isHost = id === hostId
  const x = Math.min(menu.x, innerWidth - 170)
  const y = Math.min(menu.y, innerHeight - 250)

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
          {...a11y(t(ignored ? 'context.unignore' : 'context.ignore'), 'menuitem')}
          onClick={() => app.setState((s) => ({ ignored: { ...s.ignored, [id!]: !s.ignored[id!] } }))}
        >
          {t(ignored ? 'context.unignore' : 'context.ignore')}
        </span>
      </li>
      {items.map((item) => (
        <Item key={item.id} item={item} member={m} />
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
        hosting &&
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
