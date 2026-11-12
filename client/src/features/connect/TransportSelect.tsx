// The connection method picker: auto (WebRTC, then the websocket if WebRTC cannot connect), or one
// forced. Shown on the connect screen as well as in the settings, because a stream that cannot
// connect would otherwise keep the user from ever reaching the settings to change it.
//
// The transport is chosen when the client is made, so a change takes effect on reload.
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { useNeko } from '@/state/provider'
import { TRANSPORT_MODES, type TransportMode } from '@/state/app'
import { setSetting } from '@/state/settings'
import { cn } from '@/lib/utils'

export function TransportSelect({ className }: { className?: string }) {
  const { t } = useTranslation()
  const neko = useNeko()
  const mode = useStore(neko.app, (s) => s.settings.transport)

  const change = (next: TransportMode) => {
    if (next === mode) return
    setSetting(neko, 'transport', next)
    // the client was built for the old mode; reload to build it for the new one. The URL keeps
    // the choice explicit, so it holds even before the stored setting is read.
    const url = new URL(location.href)
    url.searchParams.set('transport', next)
    location.assign(url)
  }

  return (
    <select
      aria-label={t('setting.transport')}
      title={t('setting.transport_reload')}
      data-testid="transport-select"
      className={cn(
        'cursor-pointer appearance-none rounded-[5px] border-0 bg-background-tertiary px-2 py-1.5 text-[12px] leading-5 text-text-normal [&_option]:bg-background-tertiary',
        className,
      )}
      value={mode}
      onChange={(e) => change(e.target.value as TransportMode)}
    >
      {TRANSPORT_MODES.map((m) => (
        <option key={m} value={m}>
          {t(`setting.transport_${m}`)}
        </option>
      ))}
    </select>
  )
}
