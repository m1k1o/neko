import { useEffect, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { setSetting } from '@/state/settings'
import { Banned } from './Banned'
import './settings.scss'

export function Settings() {
  const { t } = useTranslation()
  const {
    settings: s,
    broadcast,
    openInApp,
    keyboardLayouts,
  } = useStore(
    app,
    useShallow((a) => ({
      settings: a.settings,
      broadcast: a.broadcast,
      openInApp: a.openInApp,
      keyboardLayouts: a.keyboardLayouts,
    })),
  )
  const admin = useStore(client.store, selectIsAdmin)
  const [url, setUrl] = useState(broadcast.url)
  useEffect(() => setUrl(broadcast.url), [broadcast.url])

  const toggle = (
    key: 'scroll_invert' | 'autoplay' | 'ignore_emotes' | 'chat_sound' | 'links_in_app',
    label: string,
  ) => (
    <li>
      <span>{t(`setting.${label}`)}</span>
      <label className="switch">
        <input
          type="checkbox"
          aria-label={t(`setting.${label}`)}
          checked={s[key]}
          onChange={(e) => setSetting(key, e.target.checked)}
        />
        <span />
      </label>
    </li>
  )

  return (
    <div className="side-settings">
      <ul>
        <li>
          <span>{t('setting.scroll')}</span>
          <label className="slider">
            <input
              type="range"
              min="-5"
              max="5"
              step="1"
              value={s.scroll_sensitivity}
              onChange={(e) => setSetting('scroll_sensitivity', Number(e.target.value))}
            />
          </label>
        </li>
        {toggle('scroll_invert', 'scroll_invert')}
        {toggle('autoplay', 'autoplay')}
        {toggle('ignore_emotes', 'ignore_emotes')}
        {toggle('chat_sound', 'chat_sound')}
        {openInApp && toggle('links_in_app', 'links_in_app')}
        <li>
          <span>{t('setting.keyboard_layout')}</span>
          <label className="select">
            <select value={s.keyboard_layout} onChange={(e) => setSetting('keyboard_layout', e.target.value)}>
              {Object.entries(keyboardLayouts).map(([code, name]) => (
                <option key={code} value={code}>
                  {name}
                </option>
              ))}
            </select>
            <span />
          </label>
        </li>
        {admin && (
          <li className="broadcast">
            <div>
              <span>{t('setting.broadcast_title')}</span>
              {!broadcast.active ? (
                <button aria-label={t('setting.broadcast_title')} onClick={() => actions.broadcastStart(url)}>
                  <i className="fas fa-play"></i>
                </button>
              ) : (
                <button aria-label={t('setting.broadcast_title')} onClick={actions.broadcastStop} className="btn-red">
                  <i className="fas fa-stop"></i>
                </button>
              )}
            </div>
            <input
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              disabled={broadcast.active}
              className="input"
              placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
            />
          </li>
        )}
        {admin && <Banned />}
        <li>
          <button onClick={actions.logout}>{t('logout')}</button>
        </li>
      </ul>
    </div>
  )
}
