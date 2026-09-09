import * as assert from 'node:assert/strict'
import { classifyNetworkQuality } from '../src/sdk/network-monitor'
import { validateSignalingMessage } from '../src/sdk/signaling'

assert.equal(classifyNetworkQuality(null, 0, false), 'unknown')
assert.equal(classifyNetworkQuality(80, 0.01, true), 'good')
assert.equal(classifyNetworkQuality(200, 0.01, true), 'fair')
assert.equal(classifyNetworkQuality(400, 0, true), 'poor')
assert.equal(classifyNetworkQuality(80, 0.1, true), 'poor')

assert.deepEqual(validateSignalingMessage({ event: 'system/init' }), { event: 'system/init' })
assert.deepEqual(validateSignalingMessage({ event: 'session/cursors', payload: [] }), {
  event: 'session/cursors',
  payload: [],
})
assert.throws(
  () => validateSignalingMessage({ event: 'chat/message', content: 'legacy flat payload' }),
  /unsupported top-level fields/,
)
assert.throws(() => validateSignalingMessage({ event: '   ' }), /event is required/)
assert.throws(() => validateSignalingMessage({ event: 'system/init', payload: null }), /payload must be an object/)

console.log('SDK contract tests passed')
