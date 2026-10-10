import { useRef, type ComponentProps, type ReactNode } from 'react'
import * as MenuPrimitive from '@radix-ui/react-menu'
import { cn } from '@/lib/utils'

// A menu opened at a point rather than from a trigger element, on the Radix menu primitive that
// shadcn's context menu wraps: the member menu is one menu opened from the member list, a chat
// author and the keyboard, so it anchors to the coordinates it is given, as Radix's own context
// menu does with the pointer. Not modal: a right-click on another member while it is open
// opens that member's menu.
function ContextMenu({
  x,
  y,
  open,
  onOpenChange,
  children,
}: {
  x: number
  y: number
  open: boolean
  onOpenChange: (open: boolean) => void
  children: ReactNode
}) {
  const anchor = useRef({ getBoundingClientRect: () => DOMRect.fromRect({ x, y }) })
  anchor.current.getBoundingClientRect = () => DOMRect.fromRect({ x, y })
  return (
    <MenuPrimitive.Root open={open} onOpenChange={onOpenChange} modal={false}>
      <MenuPrimitive.Anchor virtualRef={anchor} />
      {children}
    </MenuPrimitive.Root>
  )
}

function ContextMenuContent({ className, ...props }: ComponentProps<typeof MenuPrimitive.Content>) {
  return (
    <MenuPrimitive.Portal>
      <MenuPrimitive.Content
        data-slot="context-menu-content"
        side="right"
        align="start"
        className={cn(
          'z-[1500] max-h-[calc(100vh-50px)] min-w-[150px] overflow-y-auto rounded-[3.5px] bg-background-floating bg-clip-padding p-[5px] text-interactive-normal shadow-elevation-high outline-hidden select-none',
          className,
        )}
        {...props}
      />
    </MenuPrimitive.Portal>
  )
}

function ContextMenuItem({ className, ...props }: ComponentProps<typeof MenuPrimitive.Item>) {
  return (
    <MenuPrimitive.Item
      data-slot="context-menu-item"
      className={cn(
        'block cursor-pointer rounded-[3px] p-[5px] font-normal whitespace-nowrap outline-hidden data-highlighted:bg-background-modifier-hover data-highlighted:text-interactive-hover',
        className,
      )}
      {...props}
    />
  )
}

function ContextMenuLabel({ className, ...props }: ComponentProps<typeof MenuPrimitive.Label>) {
  return (
    <MenuPrimitive.Label
      data-slot="context-menu-label"
      className={cn('flex flex-row gap-[5px] py-[5px]', className)}
      {...props}
    />
  )
}

function ContextMenuSeparator({ className, ...props }: ComponentProps<typeof MenuPrimitive.Separator>) {
  return (
    <MenuPrimitive.Separator
      data-slot="context-menu-separator"
      className={cn('my-[3px] h-px bg-background-secondary', className)}
      {...props}
    />
  )
}

export { ContextMenu, ContextMenuContent, ContextMenuItem, ContextMenuLabel, ContextMenuSeparator }
