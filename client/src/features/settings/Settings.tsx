import { useEffect, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { Play, Square } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import { useNeko } from '@/state/provider'
import { setSetting } from '@/state/settings'
import { Button } from '@/components/ui/button'
import { Banned } from './Banned'
import { row, label } from './classes'

const field =
  'block h-8 rounded-[5px] border border-transparent bg-background-tertiary leading-7.5 font-light text-ellipsis text-white'

export function Settings() {
  const { t } = useTranslation()
  const neko = useNeko()
  const { client, app } = neko
  const actions = useActions()
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
    name: string,
  ) => (
    <li className={row}>
      <span className={label}>{t(`setting.${name}`)}</span>
      {/* the switch: the input, its track and its knob */}
      <label className="relative h-6 w-[42px]">
        <input
          type="checkbox"
          className="peer h-0 w-0 opacity-0"
          aria-label={t(`setting.${name}`)}
          checked={s[key]}
          onChange={(e) => setSetting(neko, key, e.target.checked)}
        />
        <span className="absolute inset-0 cursor-pointer rounded-[34px] bg-background-tertiary transition-all duration-200 peer-checked:bg-style-primary" />
        <span className="absolute bottom-[3px] left-[3px] h-[18px] w-[18px] cursor-pointer rounded-full bg-white shadow-[0_2px_4px_rgba(0,0,0,0.3)] transition-all duration-300 peer-checked:translate-x-[18px]" />
      </label>
    </li>
  )

  return (
    <div className="flex flex-1" data-testid="settings">
      <ul className="flex flex-1 flex-col px-5 py-[5px]">
        <li className={row}>
          <span className={label}>{t('setting.scroll')}</span>
          <label className="max-w-[120px] whitespace-nowrap">
            <input
              type="range"
              className="slider inline-block h-6 max-w-[120px]"
              min="-5"
              max="5"
              step="1"
              value={s.scroll_sensitivity}
              onChange={(e) => setSetting(neko, 'scroll_sensitivity', Number(e.target.value))}
            />
          </label>
        </li>
        {toggle('scroll_invert', 'scroll_invert')}
        {toggle('autoplay', 'autoplay')}
        {toggle('ignore_emotes', 'ignore_emotes')}
        {toggle('chat_sound', 'chat_sound')}
        {openInApp && toggle('links_in_app', 'links_in_app')}
        <li className={row}>
          <span className={label}>{t('setting.keyboard_layout')}</span>
          <label className="max-w-[120px] text-right">
            <select
              className={cn(
                field,
                'w-full max-w-full cursor-pointer appearance-none pr-[5px] pl-2.5 text-right text-[12px] hover:border-background-secondary [&_option]:bg-background-tertiary [&_option]:font-normal [&_option]:text-text-normal',
              )}
              value={s.keyboard_layout}
              onChange={(e) => setSetting(neko, 'keyboard_layout', e.target.value)}
            >
              {Object.entries(keyboardLayouts).map(([code, name]) => (
                <option key={code} value={code}>
                  {name}
                </option>
              ))}
            </select>
          </label>
        </li>
        {admin && (
          <li className={cn(row, 'flex-col')}>
            <div className="mb-2.5 flex justify-between">
              <span className={label}>{t('setting.broadcast_title')}</span>
              {!broadcast.active ? (
                <Button
                  className="h-7.5 shrink px-2.5 py-0"
                  aria-label={t('setting.broadcast_title')}
                  onClick={() => actions.broadcastStart(url)}
                >
                  <Play className="size-3.5" />
                </Button>
              ) : (
                <Button
                  className="h-7.5 shrink bg-[#a62626] px-2.5 py-0"
                  aria-label={t('setting.broadcast_title')}
                  onClick={actions.broadcastStop}
                >
                  <Square className="size-3.5" />
                </Button>
              )}
            </div>
            <input
              className={cn(
                field,
                'px-2.5 text-left select-auto selection:bg-text-normal placeholder:text-[#757575] disabled:bg-transparent',
              )}
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              disabled={broadcast.active}
              placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
            />
          </li>
        )}
        {admin && <Banned />}
        <li className={row}>
          <Button className="my-[5px] w-full" data-testid="logout" onClick={actions.logout}>
            {t('logout')}
          </Button>
        </li>
      </ul>
    </div>
  )
}
