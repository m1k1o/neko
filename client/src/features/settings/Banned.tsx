import { useEffect, useState } from 'react'
import type { MemberData } from '@m1k1o/neko'
import { useNeko } from '@/state/hooks'
import { actions } from '@/state/actions'
import { t } from '@/i18n'

// accounts with can_login=false (see actions.ban); empty unless the provider stores accounts
export function Banned() {
  const [banned, setBanned] = useState<MemberData[] | null>(null)
  const { app } = useNeko()
  const load = () => actions.members().then((all) => setBanned(all.filter((m) => m.profile?.can_login === false)))
  useEffect(() => void load(), [app.bans]) // reloads after a ban or unban from anywhere

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
