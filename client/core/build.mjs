// After `tsc`: declaration files keep the `.ts` import extensions the sources use (tsc only
// rewrites them in the JavaScript output), and the vendored keyboard library is plain JS.
import { readdirSync, readFileSync, writeFileSync, mkdirSync, copyFileSync } from 'node:fs'

for (const f of readdirSync('dist').filter((f) => f.endsWith('.d.ts'))) {
  const src = readFileSync(`dist/${f}`, 'utf8')
  writeFileSync(
    `dist/${f}`,
    src.replace(/(from '\.\/[^']+)\.ts'/g, "$1.js'").replace(/(import\("\.\/[^"]+)\.ts"\)/g, '$1.js")'),
  )
}

mkdirSync('dist/keyboard', { recursive: true })
for (const f of ['guacamole.js', 'guacamole.d.ts']) copyFileSync(`src/keyboard/${f}`, `dist/keyboard/${f}`)
