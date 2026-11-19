// Unit tests of the chat store: `npm test` (vitest)
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { createChatStore, push } from './store.ts'

const chat = createChatStore()

const line = (content: string) => ({ id: 's1', name: 'A', type: 'text' as const, content, created: new Date(0) })

test('push appends a line with an increasing seq as a new array, and keeps the last 1000', () => {
  const before = chat.getState().lines
  let notified = 0
  const off = chat.subscribe(() => notified++)
  push(chat, line('a'))
  const after = chat.getState().lines
  assert.notEqual(after, before, 'a new array, the old one untouched')
  assert.deepEqual([before.length, after.length, notified], [0, 1, 1])
  assert.deepEqual(after[0], { ...line('a'), seq: 1 })
  chat.setState({ texts: 1 }) // another field: the lines are not replaced
  assert.equal(chat.getState().lines, after)
  for (let i = 2; i <= 1001; i++) push(chat, line(String(i)))
  const { lines } = chat.getState()
  assert.equal(lines.length, 1000)
  assert.deepEqual([lines[0].content, lines[0].seq, lines.at(-1)!.content, lines.at(-1)!.seq], ['2', 2, '1001', 1001])
  off()
})
