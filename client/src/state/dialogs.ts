// modal dialogs (legacy SweetAlert look, components/Dialog) and toasts (sonner, components/ui/sonner)
import { toast as sonner } from 'sonner'
import type { AppStore } from './app'

export interface Dialog {
  title: string
  text?: string
  icon: 'warning' | 'error' | 'info'
  cancel: boolean
  resolve: (ok: boolean) => void
}

// ask() resolves true on confirm, tell() when dismissed
function dialog(app: AppStore, d: Omit<Dialog, 'resolve'>) {
  app.getState().dialog?.resolve(false)
  return new Promise<boolean>((resolve) => app.setState({ dialog: { ...d, resolve } }))
}
export const ask = (app: AppStore, title: string, text?: string) =>
  dialog(app, { title, text, icon: 'warning', cancel: true })
export const tell = (app: AppStore, title: string, text?: string, icon: Dialog['icon'] = 'error') =>
  dialog(app, { title, text, icon, cancel: false })

let idSeq = 0
// unique ids for short-lived list items (uploads)
export const nextId = () => ++idSeq

export type ToastKind = 'info' | 'success' | 'warning' | 'error'
// a toast of a kind for 5 s, like the legacy client's
export function toast(title: string, text?: string, kind: ToastKind = 'info') {
  sonner[kind](title, { description: text, duration: 5000 })
}
export const dismissToasts = () => sonner.dismiss()
