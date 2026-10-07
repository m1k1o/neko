import type { ComponentProps } from 'react'
import './context.scss'

// the floating menu (member actions, emote picker); the caller positions it and lists the items
export function ContextMenu({ className, ...props }: ComponentProps<'ul'>) {
  return <ul className={className ? `context ${className}` : 'context'} {...props} />
}
