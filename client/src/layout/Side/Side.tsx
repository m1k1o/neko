import { useEffect } from 'react'
import { useNeko } from '@/state/hooks'
import { client, isLocked } from '@/state/client'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { t } from '@/i18n'
import { Settings } from '@/features/settings'
import { Chat } from '@/plugins/chat'
import { Files } from '@/plugins/filetransfer'
import './side.scss'

const TABS = [
  ['chat', 'fa-comment-alt'],
  ['files', 'fa-file'],
  ['settings', 'fa-sliders-h'],
] as const

export function Side() {
  const { app } = useNeko()
  const admin = client.isAdmin
  const f = app.files
  const filesAllowed =
    !!f?.enabled && (admin || !isLocked('file_transfer')) && (admin || f.user_download || f.user_upload)
  const tab = app.tab === 'files' && !filesAllowed ? 'chat' : app.tab

  useEffect(() => {
    document.querySelector('aside')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

  return (
    <aside className="neko-menu">
      <div className="tabs-container">
        <ul>
          {TABS.filter(([id]) => id !== 'files' || filesAllowed).map(([id, icon]) => (
            <li
              key={id}
              className={tab === id ? 'active' : ''}
              {...a11y(t(`side.${id}`), 'tab')}
              aria-selected={tab === id}
              onClick={() => ((app.tab = id), remember('tab', id))}
            >
              <i className={`fas ${icon}`} />
              <span>{t(`side.${id}`)}</span>
            </li>
          ))}
        </ul>
      </div>
      <div className="page-container">
        {tab === 'chat' && <Chat />}
        {tab === 'files' && <Files />}
        {tab === 'settings' && <Settings />}
      </div>
    </aside>
  )
}
