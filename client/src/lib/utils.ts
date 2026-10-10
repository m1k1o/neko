import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

// class names for a component: conditionals as clsx takes them, Tailwind conflicts resolved by the
// last one (`cn('p-2', props.className)` lets a caller override the padding)
export const cn = (...inputs: ClassValue[]) => twMerge(clsx(inputs))

// a key for a short-lived list item (an upload, an emote animation): 64 random bits from the browser's
// crypto, which every origin has (crypto.randomUUID() needs a secure context, and plain http on a LAN is not one)
export const uid = () =>
  Array.from(crypto.getRandomValues(new Uint8Array(8)), (b) => b.toString(16).padStart(2, '0')).join('')
