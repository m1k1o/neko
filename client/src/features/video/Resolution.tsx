import { useStore } from 'zustand'
import { cn } from '@/lib/utils'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'

const item =
  'relative flex cursor-pointer flex-row rounded-[3px] p-2 hover:bg-background-modifier-hover hover:text-interactive-hover focus:bg-background-modifier-hover focus:text-interactive-hover focus:outline-0'

// the screen sizes the server offers; inside the player (not fixed), so an open side panel does not cover it
export function Resolution({ onPick }: { onPick: () => void }) {
  const { size, configurations } = useStore(client.store, (s) => s.screen)
  const { width, height, rate } = size
  return (
    <ul
      className="absolute z-[1500] block max-h-[calc(100%-50px)] min-w-[150px] overflow-y-auto rounded-[3.5px] bg-background-floating bg-clip-padding p-[5px] text-interactive-normal shadow-elevation-high select-none [scrollbar-color:var(--color-background-secondary)_transparent] [scrollbar-width:thin]"
      role="menu"
      data-testid="resolution-menu"
      style={{ top: 50, right: 50 }}
      onClick={(e) => e.stopPropagation()}
    >
      {configurations.map((c, i) => {
        const active = c.width === width && c.height === height && c.rate === rate
        return (
          <li
            key={i}
            className={cn(item, active && 'bg-background-modifier-hover text-interactive-hover')}
            data-testid="resolution-item"
            data-active={active || undefined}
            {...a11y(`${c.width}x${c.height}@${c.rate}`, 'menuitem')}
            onClick={() => (client.setScreenSize(c.width, c.height, c.rate), onPick())}
          >
            <i className="fas fa-desktop mr-2.5"></i>
            <span className="grow">
              {c.width}x{c.height}
            </span>
            <small className="self-end text-[0.7em]">{c.rate}</small>
          </li>
        )
      })}
    </ul>
  )
}
