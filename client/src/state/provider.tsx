// The instance for the components below: `<NekoProvider neko={createNekoApp()}><App /></NekoProvider>`;
// a component reads it with the hooks (`useClient().send(...)`, `useStore(useApp(), (s) => s.side)`)
import { createContext, useContext, type ReactNode } from 'react'
import type { NekoApp } from './app'

const NekoContext = createContext<NekoApp | null>(null)

export function NekoProvider({ neko, children }: { neko: NekoApp; children: ReactNode }) {
  return <NekoContext value={neko}>{children}</NekoContext>
}

export function useNeko(): NekoApp {
  const neko = useContext(NekoContext)
  if (!neko) throw new Error('useNeko: no <NekoProvider> above this component')
  return neko
}
export const useClient = () => useNeko().client
export const useApp = () => useNeko().app
