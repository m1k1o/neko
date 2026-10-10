import { useEffect } from 'react'
import { useStore } from 'zustand'
import { app } from '@/state/app'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { t } from '@/i18n'
import { Settings } from '@/features/settings'
import { tabs as pluginTabs } from '@/plugins'
import type { PluginTab } from '@/plugins/types'
import { EachHook } from '@/components/EachHook'
import './side.scss'

// the plugins' tabs, then settings
const settingsTab: PluginTab = { id: 'settings', icon: 'fa-sliders-h', component: Settings }
const allTabs = [...pluginTabs(), settingsTab]
const always = () => true

export function Side() {
  useEffect(() => {
    document.querySelector('aside')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

  return (
    <EachHook items={allTabs} use={(tab) => tab.useVisible ?? always}>
      {(visible) => <Panel tabs={allTabs.filter((_, i) => visible[i])} />}
    </EachHook>
  )
}

// the tabs available now
function Panel({ tabs }: { tabs: PluginTab[] }) {
  const wanted = useStore(app, (s) => s.tab)
  // a remembered tab that is not available now falls back to the first one
  const tab = tabs.some((x) => x.id === wanted) ? wanted : tabs[0].id
  const Page = tabs.find((x) => x.id === tab)!.component

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
              onClick={() => (app.setState({ tab: id }), remember('tab', id))}
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
