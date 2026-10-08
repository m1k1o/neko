// Unit tests of the store and the emitter: `npm test` (node --test core/src/*.test.ts)
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { Store, Emitter } from './store.ts'

const fresh = () => {
  const s = new Store({ a: { b: 1 }, list: [] as number[], map: {} as Record<string, { n: number }> })
  let notified = 0
  s.subscribe(() => notified++)
  return { s, notified: () => notified }
}

test('writes anywhere in the tree are batched into one flush; watchers get new and old', async () => {
  const { s, notified } = fresh()
  const seen: unknown[] = []
  const unwatch = s.watch(
    () => s.state.a.b,
    (v, old) => seen.push([v, old]),
  )
  s.state.a.b = 2
  s.state.list.push(1)
  s.state.map.x = { n: 1 }
  assert.equal(notified(), 0)
  await Promise.resolve()
  assert.deepEqual([notified(), s.version, seen], [1, 1, [[2, 1]]])
  unwatch()
  s.state.a.b = 3
  await Promise.resolve()
  assert.deepEqual([notified(), seen.length], [2, 1])
})

test('child proxies are stable, deletes notify, no-op writes do not', async () => {
  const { s, notified } = fresh()
  assert.equal(s.state.a, s.state.a)
  s.state.a.b = 1
  await Promise.resolve()
  assert.equal(notified(), 0)
  s.state.map.x = { n: 1 }
  await Promise.resolve()
  delete s.state.map.x
  await Promise.resolve()
  assert.equal(notified(), 2)
})

test('values read from the store and put back are not wrapped again', () => {
  const { s } = fresh()
  s.state.map.y = s.state.a as any // re-assigning a proxy stores the raw object
  assert.equal(s.state.map.y, s.state.a)
  s.state.map.z = { n: 2 }
  const z = s.state.map.z
  s.state.map = { ...s.state.map }
  assert.equal(s.state.map.z, z)
  for (let i = 0; i < 3000; i++) s.state.map = { ...s.state.map }
  assert.equal(s.state.map.z.n, 2) // would overflow the stack with one proxy layer per round trip
})

test('emitter: on returns the unsubscribe, once fires one time', () => {
  const e = new Emitter<{ x: (v: number) => void }>()
  const got: number[] = []
  const off = e.on('x', (v) => got.push(v))
  e.once('x', (v) => got.push(v * 10))
  e.emit('x', 1)
  off()
  e.emit('x', 2)
  assert.deepEqual(got, [1, 10])
})
