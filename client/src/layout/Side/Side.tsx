import { useEffect } from 'react'
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { app } from '@/state/app'
import { remember } from '@/state/storage'
import { a11y } from '@/components/a11y'
import { Settings } from '@/features/settings'
import { useSlot } from '@/plugins'
import type { PluginTab } from '@/plugins/types'
import { EachHook } from '@/components/EachHook'

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
    <aside
      className="tablet:w-full tablet:portrait:h-[60vh] tablet:landscape:h-screen flex max-h-full w-side max-w-full shrink-0 flex-col bg-background-primary"
      data-testid="side"
    >
      <div className="flex h-menu max-h-full max-w-full shrink-0 bg-background-tertiary" data-testid="tabs">
        <ul className="inline-block pt-4">
          {tabs.map(({ id, icon, label }) => (
            <li
              key={id}
              className={cn(
                'mr-1 inline-block cursor-pointer rounded-t-[3px] px-2.5 py-[5px] font-semibold',
                tab === id ? 'bg-background-primary' : 'bg-background-secondary',
              )}
              {...a11y(t(label), 'tab')}
              aria-selected={tab === id}
              onClick={() => (app.setState({ tab: id }), remember('tab', id))}
            >
              <i className={`fas ${icon} mr-1 text-[10px]`} />
              <span>{t(label)}</span>
            </li>
          ))}
        </ul>
      </div>
      <div className="flex max-h-full grow overflow-auto pt-[5px]">
        <Page />
      </div>
    </aside>
  )
}
