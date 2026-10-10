import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import { langs, setLang, type Lang } from '@/i18n'

// about, the admin badge and the language picker, left of the controls
export function RoomMenu() {
  const { t, i18n } = useTranslation()
  const admin = useStore(client.store, selectIsAdmin)
  return (
    <ul className="ml-2.5 self-center">
      <li className="mr-2.5 inline-block">
        <i
          className="fas fa-question-circle cursor-pointer text-[24px]"
          data-testid="about-open"
          {...a11y('About n.eko')}
          onClick={() => app.setState({ about: true })}
        />
      </li>
      <li className="mr-2.5 inline-block">
        {admin && (
          <i
            className="fas fa-shield-alt cursor-pointer text-[24px]"
            data-testid="admin-badge"
            title={t('admin_loggedin')}
          />
        )}
      </li>
      <li className="mr-2.5 inline-block">
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
