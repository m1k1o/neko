import { useStore } from 'zustand'
import { app } from '@/state/app'

// toast kinds, the palette of the legacy vue-notification theme
const KIND = {
  info: 'border-[#187fe7] bg-[#44a4fc]',
  success: 'border-[#42a85f] bg-[#68cd86]',
  warning: 'border-[#f48a06] bg-[#ffb648]',
  error: 'border-[#b82e24] bg-[#e54d42]',
}

export function Toasts() {
  const toasts = useStore(app, (s) => s.toasts)
  return (
    <div className="pointer-events-none fixed top-[50px] left-0 z-[5000] w-[300px]">
      {toasts.map((n) => (
        <div
          key={n.id}
          className={`mx-[5px] mb-[5px] border-l-[5px] p-2.5 text-[12px] text-white ${KIND[n.kind]}`}
          data-testid="toast"
          role="status"
        >
          <div className="font-semibold">{n.title}</div>
          {n.text && <div>{n.text}</div>}
        </div>
      ))}
    </div>
  )
}
