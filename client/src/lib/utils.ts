import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

// class names for a component: conditionals as clsx takes them, Tailwind conflicts resolved by the
// last one (`cn('p-2', props.className)` lets a caller override the padding)
export const cn = (...inputs: ClassValue[]) => twMerge(clsx(inputs))
