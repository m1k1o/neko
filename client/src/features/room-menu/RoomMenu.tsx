import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import { t, langs, setLang } from '@/i18n'
import './menu.scss'

// about, the admin badge and the language picker, left of the controls
export function RoomMenu() {
  const admin = useStore(client.store, selectIsAdmin)
  const lang = useStore(app, (s) => s.lang)
  return (
    <ul className="room-settings">
      <li>
        <i className="fas fa-question-circle" {...a11y('About n.eko')} onClick={() => app.setState({ about: true })} />
      </li>
      <li>{admin && <i className="fas fa-shield-alt" title={t('admin_loggedin')} />}</li>
      <li>
        <select value={lang} onChange={(e) => setLang(e.target.value as typeof lang)} aria-label="Language">
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
