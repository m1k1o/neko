import { useNeko } from '@/state/hooks'
import './toasts.scss'

export function Toasts() {
  const { app } = useNeko()
  return (
    <div className="toasts">
      {app.toasts.map((n) => (
        <div key={n.id} className={`toast ${n.kind}`} role="status">
          <div className="toast-title">{n.title}</div>
          {n.text && <div className="toast-text">{n.text}</div>}
        </div>
      ))}
    </div>
  )
}
