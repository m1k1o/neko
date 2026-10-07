// accessibility: icon controls are <i>/<span>; this makes them focusable, labelled buttons.
// Enter/Space activation is one delegated listener (installed below) instead of per element.

// menus and pickers: a click anywhere or Escape closes them; returns the cleanup for an effect
export function closeOn(close: () => void) {
  const key = (e: KeyboardEvent) => e.key === 'Escape' && close()
  // after the click that opened it has finished propagating, or that click would close it
  const timer = setTimeout(() => {
    document.addEventListener('click', close)
    document.addEventListener('keydown', key)
  })
  return () => {
    clearTimeout(timer)
    document.removeEventListener('click', close)
    document.removeEventListener('keydown', key)
  }
}

export const a11y = (label: string, role: 'button' | 'menuitem' | 'tab' = 'button') => ({
  role,
  tabIndex: 0,
  'aria-label': label,
  title: label,
})

document.addEventListener('keydown', (e) => {
  const el = e.target as HTMLElement
  if ((e.key === 'Enter' || e.key === ' ') && /^(button|menuitem|tab)$/.test(el.getAttribute?.('role') ?? '')) {
    e.preventDefault()
    el.click()
  }
})
