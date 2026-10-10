import { useStore } from 'zustand'
import { Monitor } from 'lucide-react'
import { cn } from '@/lib/utils'
import { client } from '@/state/client'
import { DropdownMenuContent, DropdownMenuItem } from '@/components/ui/dropdown-menu'

// the screen sizes the server offers, a menu left of the icon that opens it
export function Resolution() {
  const { size, configurations } = useStore(client.store, (s) => s.screen)
  const { width, height, rate } = size
  return (
    <DropdownMenuContent
      side="left"
      align="start"
      alignOffset={-5}
      className="[scrollbar-color:var(--color-background-secondary)_transparent] [scrollbar-width:thin]"
      data-testid="resolution-menu"
    >
      {configurations.map((c, i) => {
        const active = c.width === width && c.height === height && c.rate === rate
        return (
          <DropdownMenuItem
            key={i}
            className={cn('flex flex-row p-2', active && 'bg-background-modifier-hover text-interactive-hover')}
            data-testid="resolution-item"
            data-active={active || undefined}
            onSelect={() => client.setScreenSize(c.width, c.height, c.rate)}
          >
            <Monitor className="mr-2.5 size-3.5" />
            <span className="grow">
              {c.width}x{c.height}
            </span>
            <small className="self-end text-[0.7em]">{c.rate}</small>
          </DropdownMenuItem>
        )
      })}
    </DropdownMenuContent>
  )
}
