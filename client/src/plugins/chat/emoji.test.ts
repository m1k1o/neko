// The generated emoji data (public/emoji.json, tools/emoji.ts) against its sources: `npm test`
import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const datasource = require('emoji-datasource/emoji.json') as { unified: string; short_name: string; category: string }[]
const data = JSON.parse(readFileSync(new URL('../../../public/emoji.json', import.meta.url), 'utf8')) as {
  groups: { id: string; name: string; list: string[] }[]
  list: string[]
  keywords: Record<string, string[]>
  chars: Record<string, string>
}
const charOf = (unified: string) => String.fromCodePoint(...unified.split('-').map((h) => parseInt(h, 16)))

describe('emoji.json', () => {
  // the custom emoji (tools/emoji_custom.ts) are the images of the neko group
  const images = new Set(data.groups.find((g) => g.id === 'neko')!.list)

  it('has one character per name, the one of its unified code points, and none for the custom images', () => {
    for (const name of data.list) {
      if (images.has(name)) {
        expect(data.chars[name]).toBeUndefined()
        continue
      }
      const source = datasource.find((e) => e.short_name === name)
      expect(source, name).toBeDefined()
      expect(data.chars[name]).toBe(charOf(source!.unified))
      expect([...data.chars[name]].length).toBeGreaterThan(0)
    }
    expect(Object.keys(data.chars).length).toBe(data.list.length - images.size)
  })
  it('lists every emoji of the datasource but the modifier components, once', () => {
    const names = datasource.filter((e) => e.category !== 'Component').map((e) => e.short_name)
    expect(new Set(data.list).size).toBe(data.list.length)
    expect(data.list.filter((n) => !images.has(n)).sort()).toEqual(names.sort())
  })
  it('puts every name in exactly one group and gives it keywords', () => {
    const inGroups = data.groups.flatMap((g) => g.list)
    expect(inGroups.sort()).toEqual([...data.list].sort())
    for (const name of data.list) expect(data.keywords[name]?.length, name).toBeGreaterThan(0)
  })
})
