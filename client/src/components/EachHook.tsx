import type { ReactNode } from 'react'

interface Props<I, T> {
  items: readonly I[]
  // picks an item's hook (a plugin tab's `useVisible`, `useBadge`), with the default for items without one
  use: (item: I) => () => T
  children: (values: T[]) => ReactNode
  values?: T[]
}

// Runs one hook of every item and renders `children` with the results in item order. Hooks cannot
// run in a loop, so each item's hook runs in a component of its own, nested one in the next; the
// item list must be fixed (the plugin tabs are), which keeps every component's hooks the same on
// every render.
export function EachHook<I, T>({ items, use, children, values = [] }: Props<I, T>) {
  if (!items.length) return children(values)
  return <OneHook items={items} use={use} values={values} children={children} />
}

function OneHook<I, T>({ items: [item, ...rest], use, children, values = [] }: Props<I, T>) {
  const useValue = use(item)
  const value = useValue()
  return <EachHook items={rest} use={use} values={[...values, value]} children={children} />
}
