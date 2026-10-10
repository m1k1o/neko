import { useEffect, useState } from 'react'
import type { MemberData } from '@m1k1o/neko'
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { app } from '@/state/app'

// accounts with can_login=false (see actions.ban); empty unless the provider stores accounts
export function Banned() {
  const { t } = useTranslation()
  const [banned, setBanned] = useState<MemberData[] | null>(null)
  const bans = useStore(app, (s) => s.bans)
  const load = () => actions.members().then((all) => setBanned(all.filter((m) => m.profile?.can_login === false)))
  useEffect(() => void load(), [bans]) // reloads after a ban or unban from anywhere

  if (!banned?.length) return null
  return (
    <li className="banned">
      <span>{t('setting.banned')}</span>
      <ul>
        {banned.map((m) => (
          <li key={m.id}>
            <span>{m.profile?.name || m.id}</span>
            <button onClick={() => actions.unban(m.id!)}>{t('context.unban')}</button>
          </li>
        ))}
      </ul>
    </li>
  )
}
