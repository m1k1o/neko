import type { ComponentProps } from 'react'
import { cn } from '@/lib/utils'

// the floating menu (member actions, emote picker); the caller positions it and lists the items
export function ContextMenu({ className, ...props }: ComponentProps<'ul'>) {
  return (
    <ul
      className={cn(
        'fixed z-[1500] block max-h-[calc(100%-50px)] min-w-[150px] overflow-y-auto rounded-[3.5px] bg-background-floating bg-clip-padding p-[5px] text-interactive-normal shadow-elevation-high select-none',
        className,
      )}
      {...props}
    />
  )
}
