// Fails when the import graph of src/ has a cycle (type-only imports do not count): `node tools/cycles.mjs`
import { readdirSync, readFileSync, statSync, existsSync } from 'node:fs'
import { join, dirname, resolve, relative } from 'node:path'

const root = resolve(import.meta.dirname, '../src')
const files = []
const walk = (d) => {
  for (const f of readdirSync(d)) {
    const p = join(d, f)
    if (statSync(p).isDirectory()) walk(p)
    else if (/\.tsx?$/.test(f) && !f.endsWith('.test.ts')) files.push(p)
  }
}
walk(root)

// a specifier -> the file it means, or null for packages and assets
const resolveImport = (from, spec) => {
  let base
  if (spec.startsWith('@/')) base = join(root, spec.slice(2))
  else if (spec.startsWith('.')) base = resolve(dirname(from), spec)
  else return null
  for (const c of [base, base + '.ts', base + '.tsx', join(base, 'index.ts'), join(base, 'index.tsx')])
    if (existsSync(c) && statSync(c).isFile() && /\.tsx?$/.test(c)) return c
  return null
}
const edges = new Map(
  files.map((f) => [
    f,
    [
      ...readFileSync(f, 'utf8').matchAll(
        /^(?:import|export)\s+(?!type\b)[^'"]*?from\s+['"]([^'"]+)['"]|^import\s+['"]([^'"]+)['"]/gm,
      ),
    ]
      .map((m) => resolveImport(f, m[1] ?? m[2]))
      .filter(Boolean),
  ]),
)

const state = new Map() // 1 = on the current path, 2 = done
const stack = []
let cycles = 0
const visit = (f) => {
  if (state.get(f) === 2) return
  if (state.get(f) === 1) {
    cycles++
    console.log('cycle: ' + [...stack.slice(stack.indexOf(f)), f].map((p) => relative(root, p)).join(' -> '))
    return
  }
  state.set(f, 1)
  stack.push(f)
  for (const to of edges.get(f)) visit(to)
  stack.pop()
  state.set(f, 2)
}
files.forEach(visit)
console.log(cycles ? `${cycles} import cycle(s)` : `no import cycles in ${files.length} files`)
process.exit(cycles ? 1 : 0)
