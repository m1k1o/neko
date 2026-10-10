import { cn } from '@/lib/utils'
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
    <li className="mr-2.5 inline-block">
      <i
        className={cn(
          `fas ${icon} block h-7.5 w-7.5 cursor-pointer rounded-[3px] text-center leading-8`,
          !admin && 'cursor-default opacity-80',
          locked && 'text-style-error/50',
        )}
        data-testid={`lock-${id}`}
        data-locked={locked || undefined}
        {...a11y(tip)}
        aria-disabled={!admin}
        onClick={() => admin && onToggle()}
      />
    </li>
  )
}
