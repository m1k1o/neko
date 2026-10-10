import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { CircleHelp, Shield } from 'lucide-react'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { IconButton } from '@/components/IconButton'
import { langs, setLang, type Lang } from '@/i18n'

// about, the admin badge and the language picker, left of the controls
export function RoomMenu() {
  const { t, i18n } = useTranslation()
  const admin = useStore(client.store, selectIsAdmin)
  return (
    <ul className="ml-2.5 self-center">
      <li className="mr-2.5 inline-block align-middle">
        <IconButton label="About n.eko" data-testid="about-open" onClick={() => app.setState({ about: true })}>
          <CircleHelp className="size-6" />
        </IconButton>
      </li>
      <li className="mr-2.5 inline-block align-middle">
        {admin && (
          <span className="inline-block" data-testid="admin-badge" title={t('admin_loggedin')}>
            <Shield className="size-6" role="img" aria-label={t('admin_loggedin')} />
          </span>
        )}
      </li>
      <li className="mr-2.5 inline-block align-middle">
        <select
          className="inline-block h-[26px] cursor-pointer appearance-none rounded-[5px] border border-background-primary bg-background-tertiary align-text-bottom text-white [&_option]:bg-background-tertiary [&_option]:font-normal [&_option]:text-text-normal"
          data-testid="lang"
          value={i18n.language}
          onChange={(e) => setLang(e.target.value as Lang)}
          aria-label="Language"
        >
          {langs.map((l) => (
            <option key={l} value={l}>
              {l}
            </option>
          ))}
        </select>
      </li>
    </ul>
  )
}
