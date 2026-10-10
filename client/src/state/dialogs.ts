// modal dialogs (legacy SweetAlert look, components/Dialog) and toasts (sonner, components/ui/sonner)
import { toast as sonner } from 'sonner'
import { app } from './app'

export interface Dialog {
  title: string
  text?: string
  icon: 'warning' | 'error' | 'info'
  cancel: boolean
  resolve: (ok: boolean) => void
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
// unique ids for short-lived list items (uploads)
export const nextId = () => ++idSeq

export type ToastKind = 'info' | 'success' | 'warning' | 'error'
// a toast of a kind for 5 s, like the legacy client's
export function toast(title: string, text?: string, kind: ToastKind = 'info') {
  sonner[kind](title, { description: text, duration: 5000 })
}
export const dismissToasts = () => sonner.dismiss()
