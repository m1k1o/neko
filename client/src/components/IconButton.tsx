import type { ComponentProps } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const iconButton = cva(
  'inline-flex cursor-pointer items-center justify-center focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-style-primary aria-disabled:cursor-default aria-disabled:opacity-80',
  {
    variants: {
      variant: {
        // the icon alone: the room bar, the chat, the files
        plain: '',
        // the header menu: a 30px box
        header: 'flex h-7.5 w-7.5 rounded-[3px]',
        // the menus over the video: a translucent 30px box
        video: 'h-7.5 w-7.5 rounded-[5px] bg-white/20 text-white/60',
      },
    },
    defaultVariants: { variant: 'plain' },
  },
)

// An icon control: a real button, so Enter, Space and focus come for free, named by its label
// (the accessible name and the tooltip); the icon is the child.
export function IconButton({
  label,
  variant,
  className,
  ...props
}: { label: string } & VariantProps<typeof iconButton> & Omit<ComponentProps<'button'>, 'aria-label' | 'title'>) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      className={cn(iconButton({ variant }), className)}
      {...props}
    />
  )
}
