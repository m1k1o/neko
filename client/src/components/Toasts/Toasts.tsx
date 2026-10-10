import { useStore } from 'zustand'
import { app } from '@/state/app'
import './toasts.scss'

export function Toasts() {
  const toasts = useStore(app, (s) => s.toasts)
  return (
    <div className="toasts">
      {toasts.map((n) => (
        <div key={n.id} className={`toast ${n.kind}`} role="status">
          <div className="toast-title">{n.title}</div>
          {n.text && <div className="toast-text">{n.text}</div>}
        </div>
      ))}
    </div>
  )
}
