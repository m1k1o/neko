import type { ComponentProps } from 'react'
import * as TabsPrimitive from '@radix-ui/react-tabs'
import { cn } from '@/lib/utils'

// shadcn/ui tabs on Radix (the side panel): real buttons with role=tab, arrow keys between them
function Tabs(props: ComponentProps<typeof TabsPrimitive.Root>) {
  return <TabsPrimitive.Root data-slot="tabs" {...props} />
}

function TabsList({ className, ...props }: ComponentProps<typeof TabsPrimitive.List>) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      className={cn('flex h-menu max-h-full max-w-full shrink-0 bg-background-tertiary pt-4', className)}
      {...props}
    />
  )
}

function TabsTrigger({ className, ...props }: ComponentProps<typeof TabsPrimitive.Trigger>) {
  return (
    <TabsPrimitive.Trigger
      data-slot="tabs-trigger"
      className={cn(
        'mr-1 inline-block cursor-pointer rounded-t-[3px] bg-background-secondary px-2.5 py-[5px] font-semibold focus-visible:outline-2 focus-visible:outline-style-primary data-[state=active]:bg-background-primary',
        className,
      )}
      {...props}
    />
  )
}

function TabsContent({ className, ...props }: ComponentProps<typeof TabsPrimitive.Content>) {
  return <TabsPrimitive.Content data-slot="tabs-content" className={cn('outline-hidden', className)} {...props} />
}

export { Tabs, TabsList, TabsTrigger, TabsContent }
