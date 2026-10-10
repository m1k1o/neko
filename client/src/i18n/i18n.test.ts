// Unit tests of the i18n runtime: `npm test` (vitest). The browser is faked just far enough for it:
// location (?lang=), document (<html lang>), navigator.language and localStorage.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import i18next from 'i18next'
import type { LocaleFiles } from './index.ts'

const stored = new Map<string, string>()
const fake = (name: string, value: unknown) =>
  Object.defineProperty(globalThis, name, { value, configurable: true, writable: true })
fake('location', { href: 'http://neko.test/?lang=de' })
fake('document', { documentElement: { lang: '' } })
fake('navigator', { language: 'en-US' })
fake('localStorage', {
  getItem: (k: string) => stored.get(k) ?? null,
  setItem: (k: string, v: string) => stored.set(k, v),
  removeItem: (k: string) => stored.delete(k),
})
const { i18n, t, initI18n, setLang, loader, langs, detectLang, htmlLang } = await import('./index.ts')

const NS = ['common', 'chat', 'files'] as const
const json = (lang: string, ns: string) =>
  JSON.parse(readFileSync(new URL(`../locales/${lang}/${ns}.json`, import.meta.url), 'utf8')) as object
const flat = (o: object, prefix = ''): Record<string, string> =>
  Object.fromEntries(
    Object.entries(o).flatMap(([k, v]) =>
      typeof v === 'string' ? [[prefix + k, v]] : Object.entries(flat(v, prefix + k + '.')),
    ),
  )
const vars = (s: string) =>
  [...s.matchAll(/\{\{(\w+)\}\}/g)]
    .map((m) => m[1])
    .sort()
    .join()

test('start-up: ?lang= wins and is kept, the language and en are loaded, the others are not', async () => {
  await initI18n(['chat', 'files'])
  assert.equal(i18n.language, 'de')
  assert.equal(stored.get('lang'), 'de')
  assert.equal(document.documentElement.lang, 'de')
  for (const ns of NS) {
    assert.ok(i18n.hasResourceBundle('de', ns), `de/${ns}`)
    assert.ok(i18n.hasResourceBundle('en', ns), `en/${ns}`)
  }
  for (const lang of langs) if (lang !== 'de' && lang !== 'en') assert.ok(!i18n.hasResourceBundle(lang, 'common'), lang)
})

test('namespaces: common is the default, the plugins are <ns>:key', () => {
  assert.equal(t('side.settings'), 'Einstellungen')
  assert.equal(t('chat:tab'), 'Chat')
  assert.equal(t('files:tab'), 'Dateien')
  assert.equal(t('files:locks.notif_locked'), 'Dateiübertragung gesperrt')
})

test('a key the language lacks falls back to en', () => {
  assert.equal(t('context.ban'), 'Ban')
  assert.equal(t('setting.banned'), 'Banned')
})

test('interpolation: {{x}}, count, and a value is not interpolated or nested again, nor escaped', () => {
  assert.equal(t('notifications.resolution', { width: 1, height: 2, rate: 3 }), 'die Auflösung geändert zu 1x2@3')
  assert.equal(
    t('files:delete_selected_confirm', { count: 3 }),
    'Möchten Sie die 3 ausgewählten Dateien wirklich entfernen?',
  )
  assert.equal(t('chat:muted', { name: '{{x}} $t(logout) <b>' }), '{{x}} $t(logout) <b> stummgeschaltet')
})

test('a key no language has is reported (development) and shown as the key', () => {
  const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
  assert.equal(t('nope.key'), 'nope.key')
  assert.deepEqual(warn.mock.calls, [['i18n: missing common:nope.key (en)']])
  warn.mockRestore()
})

test('setLang: loads the language, persists it, sets <html lang> (zh-TW for tw)', async () => {
  await setLang('tw')
  assert.equal(i18n.language, 'tw')
  assert.equal(stored.get('lang'), 'tw')
  assert.equal(document.documentElement.lang, 'zh-TW')
  assert.equal(t('side.settings'), '設定')
  assert.equal(t('files:tab'), '檔案')
  await setLang('en')
  assert.equal(document.documentElement.lang, 'en')
  assert.equal(t('side.settings'), 'Settings')
})

test('htmlLang maps the two Chinese codes', () => {
  assert.equal(htmlLang('cn'), 'zh-CN')
  assert.equal(htmlLang('tw'), 'zh-TW')
  assert.equal(htmlLang('de'), 'de')
})

test('detectLang: exact code, then its base, else en', () => {
  const detect = (language: string) => (fake('navigator', { language }), detectLang())
  assert.equal(detect('de-DE'), 'de')
  assert.equal(detect('sk'), 'sk')
  assert.equal(detect('EN-us'), 'en')
  assert.equal(detect('pt-BR'), 'en')
  assert.equal(detect('zh-CN'), 'en') // the Chinese locales are coded cn/tw, not zh
})

test('loader: reads exactly <lang>/<ns> from the glob map, nothing of the other languages', async () => {
  const de = vi.fn(() => Promise.resolve({ default: { a: 'A' } }))
  const fr = vi.fn(() => Promise.resolve({ default: { a: 'Á' } }))
  const files: LocaleFiles = { '../locales/de/common.json': de, '../locales/fr/common.json': fr }
  const i = i18next.createInstance().use(loader(files))
  await i.init({ lng: 'de', fallbackLng: false, ns: ['common'], defaultNS: 'common' })
  assert.equal(i.t('a'), 'A')
  assert.equal(de.mock.calls.length, 1)
  assert.equal(fr.mock.calls.length, 0)
  // a language without files: nothing loaded, no error out of changeLanguage
  await i.changeLanguage('xx')
  assert.equal(i.language, 'xx')
  assert.ok(!i.hasResourceBundle('xx', 'common'))
  assert.equal(de.mock.calls.length, 1)
})

test('every language has the three files, en every key, the others all but the known gaps', () => {
  assert.deepEqual(readdirSync(new URL('../locales/', import.meta.url)).sort(), [...langs].sort())
  const en = Object.fromEntries(NS.map((ns) => [ns, flat(json('en', ns))]))
  assert.equal(Object.keys(en.common).length + Object.keys(en.chat).length + Object.keys(en.files).length, 106)
  // strings added since the translations: en only, until someone translates them
  const GAPS = [
    'context.ban',
    'context.unban',
    'context.confirm.ban_text',
    'controls.mic_on',
    'controls.mic_off',
    'controls.mic_error',
    'setting.links_in_app',
    'setting.banned',
    'connection.replaced',
  ]
  for (const lang of langs) {
    assert.deepEqual(readdirSync(new URL(`../locales/${lang}/`, import.meta.url)).sort(), [
      'chat.json',
      'common.json',
      'files.json',
    ])
    const missing: string[] = []
    for (const ns of NS) {
      const mine = flat(json(lang, ns))
      for (const k of Object.keys(en[ns])) {
        if (!(k in mine)) missing.push(ns === 'common' ? k : `${ns}:${k}`)
        else assert.equal(vars(mine[k]), vars(en[ns][k]), `${lang} ${ns}:${k} interpolates differently`)
      }
      for (const k of Object.keys(mine)) assert.ok(k in en[ns], `${lang} has ${ns}:${k}, en has not`)
    }
    assert.deepEqual(missing, lang === 'en' ? [] : lang === 'sk' ? ['you', ...GAPS] : GAPS, lang)
  }
})

test('every key the sources use exists in en: literal keys and tab labels, the prefix of a template key', () => {
  const en = Object.fromEntries(NS.map((ns) => [ns, flat(json('en', ns))]))
  const has = (key: string, where: string, prefix = false) => {
    const [ns, k] = key.includes(':') ? key.split(':') : ['common', key]
    assert.ok(ns in en, `${where}: namespace of '${key}'`)
    assert.ok(prefix ? Object.keys(en[ns]).some((x) => x.startsWith(k)) : k in en[ns], `${where}: '${key}'`)
  }
  const walk = (dir: string): string[] =>
    readdirSync(dir, { withFileTypes: true }).flatMap((e) =>
      e.isDirectory()
        ? walk(join(dir, e.name))
        : /\.tsx?$/.test(e.name) && !e.name.endsWith('.test.ts')
          ? [join(dir, e.name)]
          : [],
    )
  let n = 0
  for (const file of walk(new URL('../', import.meta.url).pathname)) {
    const src = readFileSync(file, 'utf8')
    const where = file.slice(file.indexOf('/src/') + 5)
    // t('key'), t(cond ? 'a' : 'b', ...): every quoted key up to the closing parenthesis, templates aside
    for (const [, args] of src.matchAll(/\bt\(([^)]*)\)/g))
      for (const [, key] of args.replace(/`[^`]*`/g, '').matchAll(/'([\w:.]+)'/g)) (has(key, where), n++)
    // t(`group.${x}`): the part before the expression is the start of some key
    for (const [, prefix] of src.matchAll(/\bt\(`([\w:.]+)\$\{/g)) (has(prefix, where, true), n++)
    // a tab's label
    for (const [, key] of src.matchAll(/\blabel: '([\w:.]+)'/g)) (has(key, where), n++)
  }
  assert.ok(n > 80, `${n} keys checked`)
})
