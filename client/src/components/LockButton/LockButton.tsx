import { a11y } from '@/components/a11y'

// a lock icon in the header menu: admins toggle it, users only see its state
export function LockButton({
  icon,
  locked,
  admin,
  tip,
  onToggle,
}: {
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
        {...a11y(tip)}
        aria-disabled={!admin}
        onClick={() => admin && onToggle()}
      />
    </li>
  )
}
