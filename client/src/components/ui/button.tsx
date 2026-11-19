import type { ComponentProps } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

// shadcn/ui's button with the variants of the legacy client's text buttons
const buttonVariants = cva(
  'inline-flex cursor-pointer items-center justify-center border-0 text-center focus-visible:outline-2 focus-visible:outline-style-primary',
  {
    variants: {
      variant: {
        // connect, settings, about, the broadcast controls
        primary: 'rounded-[5px] bg-style-primary p-1 leading-7.5 font-bold text-text-normal uppercase',
        // the OAuth login
        outline:
          'rounded-[5px] border border-style-primary bg-background-tertiary p-1 leading-7.5 font-bold text-text-normal uppercase',
        // the dialog's confirm and cancel
        confirm: 'rounded-[0.25em] bg-background-tertiary px-[2em] py-[0.625em] text-[1.0625em] text-white',
        cancel: 'rounded-[0.25em] bg-background-floating px-[2em] py-[0.625em] text-[1.0625em] text-white',
      },
    },
    defaultVariants: { variant: 'primary' },
  },
)

function Button({ className, variant, ...props }: ComponentProps<'button'> & VariantProps<typeof buttonVariants>) {
  return <button data-slot="button" className={cn(buttonVariants({ variant }), className)} {...props} />
}

export { Button, buttonVariants }
