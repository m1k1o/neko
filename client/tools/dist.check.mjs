// Runnable check of the built client in dist/ (after `vite build`, part of `npm run build`): the main
// chunk carries no language; each src/locales/<lang>/<ns>.json is a chunk of its own, fetched when
// needed, named assets/locales/<lang>-<ns>-<hash>.js (vite.config.ts)
import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const assets = join(import.meta.dirname, '../dist/assets')
const chunks = readdirSync(assets, { recursive: true }).filter((f) => f.endsWith('.js'))
const main = chunks.find((f) => f.startsWith('index-'))
const read = (f) => readFileSync(join(assets, f), 'utf8')
const mainSrc = read(main)
const locales = join(import.meta.dirname, '../src/locales')
const langs = readdirSync(locales)
for (const lang of langs) {
  // a string of each language that only its files have: the settings tab label
  const label = JSON.parse(readFileSync(join(locales, lang, 'common.json'), 'utf8')).side.settings
  assert.ok(!mainSrc.includes(label), `${main} carries ${lang} ("${label}")`)
  for (const file of readdirSync(join(locales, lang))) {
    const ns = file.replace(/\.json$/, '')
    const own = chunks.filter((f) => f.startsWith(`locales/${lang}-${ns}-`))
    assert.equal(own.length, 1, `locales/${lang}-${ns}-*.js: ${own.length} chunks`)
    if (ns === 'common') assert.ok(read(own[0]).includes(label), `${own[0]} lacks ${lang} ("${label}")`)
  }
}
console.log(`dist ok: ${langs.length} languages in chunks of their own, none in ${main}`)
