import { useEffect } from 'react'
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { SlidersHorizontal } from 'lucide-react'
import { app } from '@/state/app'
import { remember } from '@/state/storage'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Settings } from '@/features/settings/Settings'
import { useSlot } from '@/plugins'
import type { PluginTab } from '@/plugins/types'
import { EachHook } from '@/components/EachHook'

// after the plugins' tabs
const settingsTab: PluginTab = { id: 'settings', icon: SlidersHorizontal, label: 'side.settings', component: Settings }
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
    <Tabs value={tab} onValueChange={(id) => (app.setState({ tab: id }), remember('tab', id))} asChild>
      <aside
        className="tablet:w-full tablet:portrait:h-[60vh] tablet:landscape:h-screen flex max-h-full w-side max-w-full shrink-0 flex-col bg-background-primary"
        data-testid="side"
      >
        <TabsList data-testid="tabs">
          {tabs.map(({ id, icon: Icon, label }) => (
            <TabsTrigger key={id} value={id}>
              <Icon className="mr-1 inline size-2.5" />
              <span>{t(label)}</span>
            </TabsTrigger>
          ))}
        </TabsList>
        <TabsContent value={tab} className="flex max-h-full grow overflow-auto pt-[5px]">
          <Page />
        </TabsContent>
      </aside>
    </Tabs>
  )
}
