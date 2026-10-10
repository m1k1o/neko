import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import { langs, setLang, type Lang } from '@/i18n'
import './menu.scss'

// about, the admin badge and the language picker, left of the controls
export function RoomMenu() {
  const { t, i18n } = useTranslation()
  const admin = useStore(client.store, selectIsAdmin)
  return (
    <ul className="room-settings ml-2.5 flex items-center justify-start">
      <li>
        <i
          className="fas fa-question-circle"
          data-testid="about-open"
          {...a11y('About n.eko')}
          onClick={() => app.setState({ about: true })}
        />
      </li>
      <li>{admin && <i className="fas fa-shield-alt" data-testid="admin-badge" title={t('admin_loggedin')} />}</li>
      <li>
        <select
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
