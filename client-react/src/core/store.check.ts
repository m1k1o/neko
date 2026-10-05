// Runnable self-check: `node src/core/store.check.ts`
import assert from 'node:assert/strict'
import { Store } from './store.ts'

const s = new Store({ a: { b: 1 }, list: [] as number[], map: {} as Record<string, { n: number }> })
let notified = 0
const seen: unknown[] = []
s.subscribe(() => notified++)
s.watch(
  () => s.state.a.b,
  (v, old) => seen.push([v, old]),
)

// nested writes are batched into one flush
s.state.a.b = 2
s.state.list.push(1)
s.state.map.x = { n: 1 }
assert.equal(notified, 0)
await Promise.resolve()
assert.equal(notified, 1)
assert.equal(s.version, 1)
assert.deepEqual(seen, [[2, 1]])

// child proxies are stable, deletes notify, no-op writes don't
assert.equal(s.state.a, s.state.a)
s.state.a.b = 2
await Promise.resolve()
assert.equal(notified, 1)
delete s.state.map.x
await Promise.resolve()
assert.equal(notified, 2)

// re-assigning a proxy stores the raw object, identity stays stable
s.state.map.y = s.state.a as any
assert.equal(s.state.map.y, s.state.a)

console.log('store ok')
