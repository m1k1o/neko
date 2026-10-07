import { useEffect } from 'react'
import { useNeko } from '@/state/hooks'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { t } from '@/i18n'
import { Settings } from '@/features/settings'
import { tabs as pluginTabs } from '@/plugins'
import type { PluginTab } from '@/plugins/types'
import './side.scss'

// the plugins' tabs, then settings
const settingsTab: PluginTab = { id: 'settings', icon: 'fa-sliders-h', component: Settings }

export function Side() {
  const { app } = useNeko()
  const tabs = [...pluginTabs(), settingsTab].filter((x) => !x.visible || x.visible())
  // a remembered tab that is not available now falls back to the first one
  const tab = tabs.some((x) => x.id === app.tab) ? app.tab : tabs[0].id
  const Page = tabs.find((x) => x.id === tab)!.component

  useEffect(() => {
    document.querySelector('aside')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

  return (
    <aside className="neko-menu">
      <div className="tabs-container">
        <ul>
          {tabs.map(({ id, icon }) => (
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
        <Page />
      </div>
    </aside>
  )
}
