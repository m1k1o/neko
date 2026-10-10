import type { CSSProperties } from 'react'
import { Toaster as Sonner } from 'sonner'

// shadcn/ui's sonner toaster, at the top left in the look of the legacy vue-notification theme:
// a list of 290px toasts with a coloured left border per kind, no icons (state/dialogs.ts toast())
export function Toaster() {
  return (
    <Sonner
      position="top-left"
      offset={{ top: 50, left: 5 }}
      mobileOffset={{ top: 50, left: 5 }}
      gap={5}
      expand
      visibleToasts={9}
      style={{ '--width': '290px' } as CSSProperties}
      icons={{ info: null, success: null, warning: null, error: null }}
      toastOptions={{
        unstyled: true,
        classNames: {
          toast: 'border-l-[5px] p-2.5 text-[12px] text-white',
          title: 'font-semibold',
          info: 'border-[#187fe7] bg-[#44a4fc]',
          success: 'border-[#42a85f] bg-[#68cd86]',
          warning: 'border-[#f48a06] bg-[#ffb648]',
          error: 'border-[#b82e24] bg-[#e54d42]',
        },
      }}
    />
  )
}
