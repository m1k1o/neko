import type { ComponentProps } from 'react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { cn } from '@/lib/utils'

// shadcn/ui dialog on Radix: focus trap, Escape, outside click and ARIA; the overlay and the
// content are composed by the caller, as their dimming and box differ per dialog
function Dialog(props: ComponentProps<typeof DialogPrimitive.Root>) {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />
}

function DialogPortal(props: ComponentProps<typeof DialogPrimitive.Portal>) {
  return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />
}

function DialogOverlay({ className, ...props }: ComponentProps<typeof DialogPrimitive.Overlay>) {
  return (
    <DialogPrimitive.Overlay
      data-slot="dialog-overlay"
      className={cn('fixed inset-0 z-50 bg-black/40', className)}
      {...props}
    />
  )
}

function DialogContent({ className, ...props }: ComponentProps<typeof DialogPrimitive.Content>) {
  return (
    <DialogPrimitive.Content
      data-slot="dialog-content"
      className={cn('fixed top-1/2 left-1/2 z-50 -translate-x-1/2 -translate-y-1/2 outline-hidden', className)}
      {...props}
    />
  )
}

function DialogTitle(props: ComponentProps<typeof DialogPrimitive.Title>) {
  return <DialogPrimitive.Title data-slot="dialog-title" {...props} />
}

function DialogDescription(props: ComponentProps<typeof DialogPrimitive.Description>) {
  return <DialogPrimitive.Description data-slot="dialog-description" {...props} />
}

export { Dialog, DialogPortal, DialogOverlay, DialogContent, DialogTitle, DialogDescription }
