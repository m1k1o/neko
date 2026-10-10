import { a11y } from '@/components/a11y'

// a lock icon in the header menu: admins toggle it, users only see its state
export function LockButton({
  id,
  icon,
  locked,
  admin,
  tip,
  onToggle,
}: {
  // the test id of the button, `lock-<id>`
  id: string
  icon: string
  locked: boolean
  admin: boolean
  tip: string
  onToggle: () => void
}) {
  return (
    <li>
      <i
        className={`fas ${icon}${admin ? '' : ' disabled'}${locked ? ' locked' : ''}`}
        data-testid={`lock-${id}`}
        data-locked={locked || undefined}
        {...a11y(tip)}
        aria-disabled={!admin}
        onClick={() => admin && onToggle()}
      />
    </li>
  )
}
