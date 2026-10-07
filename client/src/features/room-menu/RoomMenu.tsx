import { useNeko } from '@/state/hooks'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import { t, langs, setLang } from '@/i18n'
import './menu.scss'

// about, the admin badge and the language picker, left of the controls
export function RoomMenu() {
  useNeko()
  return (
    <ul className="room-settings">
      <li>
        <i className="fas fa-question-circle" {...a11y('About n.eko')} onClick={() => (app.state.about = true)} />
      </li>
      <li>{client.isAdmin && <i className="fas fa-shield-alt" title={t('admin_loggedin')} />}</li>
      <li>
        <select
          value={app.state.lang}
          onChange={(e) => setLang(e.target.value as typeof app.state.lang)}
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
