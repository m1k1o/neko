// accessibility: icon controls are <i>/<span>; this makes them focusable, labelled buttons.
// Enter/Space activation is one delegated listener (installed below) instead of per element.

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
