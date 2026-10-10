import { useEffect } from 'react'
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { Settings } from '@/features/settings'
import { useSlot } from '@/plugins'
import type { PluginTab } from '@/plugins/types'
import { EachHook } from '@/components/EachHook'
import './side.scss'

// after the plugins' tabs
const settingsTab: PluginTab = { id: 'settings', icon: 'fa-sliders-h', label: 'side.settings', component: Settings }
const always = () => true

export function Side() {
  const tabs = [...useSlot('side.tab'), settingsTab]
  useEffect(() => {
    document.querySelector('aside')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

  return (
    <EachHook items={tabs} use={(tab) => tab.useVisible ?? always}>
      {(visible) => <Panel tabs={tabs.filter((_, i) => visible[i])} />}
    </EachHook>
  )
}

// the tabs available now
function Panel({ tabs }: { tabs: PluginTab[] }) {
  const { t } = useTranslation()
  const wanted = useStore(app, (s) => s.tab)
  // a remembered tab that is not available now falls back to the first one
  const tab = tabs.some((x) => x.id === wanted) ? wanted : tabs[0].id
  const Page = tabs.find((x) => x.id === tab)!.component

  return (
    <aside className="neko-menu">
      <div className="tabs-container">
        <ul>
          {tabs.map(({ id, icon, label }) => (
            <li
              key={id}
              className={tab === id ? 'active' : ''}
              {...a11y(t(label), 'tab')}
              aria-selected={tab === id}
              onClick={() => (app.setState({ tab: id }), remember('tab', id))}
            >
              <i className={`fas ${icon}`} />
              <span>{t(label)}</span>
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
