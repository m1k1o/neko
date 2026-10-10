// modal dialogs (legacy SweetAlert look) and toasts (legacy vue-notification look)
import { app } from './app'

export interface Dialog {
  title: string
  text?: string
  icon: 'warning' | 'error' | 'info'
  cancel: boolean
  resolve: (ok: boolean) => void
}

export interface Toast {
  id: number
  kind: 'info' | 'success' | 'warning' | 'error'
  title: string
  text?: string
}

// ask() resolves true on confirm, tell() when dismissed
function dialog(d: Omit<Dialog, 'resolve'>) {
  app.getState().dialog?.resolve(false)
  return new Promise<boolean>((resolve) => app.setState({ dialog: { ...d, resolve } }))
}
export const ask = (title: string, text?: string) => dialog({ title, text, icon: 'warning', cancel: true })
export const tell = (title: string, text?: string, icon: Dialog['icon'] = 'error') =>
  dialog({ title, text, icon, cancel: false })

let idSeq = 0
// unique ids for toasts and other short-lived list items
export const nextId = () => ++idSeq

export function toast(title: string, text?: string, kind: Toast['kind'] = 'info') {
  const id = nextId()
  app.setState((s) => ({ toasts: [...s.toasts, { id, kind, title, text }] }))
  setTimeout(() => app.setState((s) => ({ toasts: s.toasts.filter((x) => x.id !== id) })), 5000)
}
