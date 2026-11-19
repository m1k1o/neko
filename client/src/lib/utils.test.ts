import { describe, expect, it } from 'vitest'
import { cn } from './utils'

describe('cn', () => {
  it('joins names and drops falsy conditionals', () => {
    expect(cn('a', false && 'b', undefined, 'c', { d: true, e: false })).toBe('a c d')
  })
  it('resolves a Tailwind conflict in favour of the last class', () => {
    expect(cn('p-2 text-text-normal', 'p-[5px]')).toBe('text-text-normal p-[5px]')
    expect(cn('bg-background-primary', 'bg-background-floating')).toBe('bg-background-floating')
  })
  it('keeps classes that do not conflict', () => {
    expect(cn('flex h-menu', 'w-side')).toBe('flex h-menu w-side')
  })
})
