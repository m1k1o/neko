import { useEffect, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin, type Session } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { useActions } from '@/state/actions'
import type { AppStore } from '@/state/app'
import { ask } from '@/state/dialogs'
import { useNeko } from '@/state/provider'
import { Avatar } from '@/components/Avatar'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuSeparator,
} from '@/components/ui/context-menu'
import type { PluginMemberMenuItem } from '@/plugins/types'

const confirmThen = (app: AppStore, title: string, text: string, fn: () => void) =>
  ask(app, title, text).then((ok) => ok && fn())
const always = () => true

// a plugin's entry, when its hook shows it for this member
function Item({ entry, member }: { entry: PluginMemberMenuItem; member: Session }) {
  const neko = useNeko()
  const useVisible = entry.useVisible ?? always
  const visible = useVisible(member)
  if (!visible) return null
  return <ContextMenuItem onSelect={() => entry.onClick(neko, member)}>{entry.label(neko, member)}</ContextMenuItem>
}

// rendered by App outside the room bar, which is hidden at narrow widths where the chat still works;
// opened at the point of a right-click (members, chat authors) or under a focused member (Shift+F10)
export function MemberMenu({ items }: { items: readonly PluginMemberMenuItem[] }) {
  const { t } = useTranslation()
  const { client, app } = useNeko()
  const actions = useActions()
  const menu = useStore(app, (s) => s.menu)
  const [bannable, setBannable] = useState(false)

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
  }, [id, client, actions])
  const open = !!menu && !!m
  const n = m?.profile.name ?? ''
  const isHost = id === hostId
  const danger = 'text-[#f04747]'

  return (
    <ContextMenu x={menu?.x ?? 0} y={menu?.y ?? 0} open={open} onOpenChange={(o) => !o && app.setState({ menu: null })}>
      {open && (
        <ContextMenuContent aria-label={n} data-testid="member-menu">
          <ContextMenuLabel>
            <Avatar seed={n} avatar={m.profile.avatar} size={25} />
            <strong className="max-w-[200px] leading-[25px] font-bold text-ellipsis">{n}</strong>
          </ContextMenuLabel>
          <ContextMenuSeparator />
          <ContextMenuItem
            onSelect={() => app.setState((s) => ({ ignored: { ...s.ignored, [id!]: !s.ignored[id!] } }))}
          >
            {t(ignored ? 'context.unignore' : 'context.ignore')}
          </ContextMenuItem>
          {items.map((item) => (
            <Item key={item.id} entry={item} member={m} />
          ))}
          {admin ? (
            <>
              {!implicit && isHost && (
                <>
                  <ContextMenuItem onSelect={actions.reset}>{t('context.release')}</ContextMenuItem>
                  <ContextMenuItem onSelect={actions.take}>{t('context.take')}</ContextMenuItem>
                </>
              )}
              {!implicit && !isHost && (
                <ContextMenuItem onSelect={() => actions.give(id!)}>{t('context.give')}</ContextMenuItem>
              )}
            </>
          ) : (
            hosting &&
            !implicit && <ContextMenuItem onSelect={() => actions.give(id!)}>{t('context.give')}</ContextMenuItem>
          )}
          {admin && !m.profile.is_admin && (
            <>
              <ContextMenuSeparator />
              <ContextMenuItem
                className={danger}
                onSelect={() =>
                  confirmThen(
                    app,
                    t('context.confirm.kick_title', { name: n }),
                    t('context.confirm.kick_text', { name: n }),
                    () => actions.kick(id!),
                  )
                }
              >
                {t('context.kick')}
              </ContextMenuItem>
              {bannable && (
                <ContextMenuItem
                  className={danger}
                  onSelect={() =>
                    confirmThen(
                      app,
                      t('context.confirm.ban_title', { name: n }),
                      t('context.confirm.ban_text', { name: n }),
                      () => actions.ban(id!),
                    )
                  }
                >
                  {t('context.ban')}
                </ContextMenuItem>
              )}
            </>
          )}
        </ContextMenuContent>
      )}
    </ContextMenu>
  )
}
