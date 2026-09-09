import * as assert from 'node:assert/strict'
import { AuthClient } from '../src/sdk/auth'
import { classifyNetworkQuality } from '../src/sdk/network-monitor'
import { validateSignalingMessage } from '../src/sdk/signaling'

async function testAuthClient() {
  const requests: Array<{ url: string; data: unknown }> = []
  const http = {
    defaults: { headers: { common: {} as Record<string, unknown> } },
    async post<T>(url: string, data?: unknown) {
      requests.push({ url, data })
      return { data: (url.endsWith('/login') ? { token: 'test-token' } : true) as T }
    },
  }
  const auth = new AuthClient(http, '/api')
  assert.equal(await auth.login('alice', 'secret'), 'test-token')
  assert.equal(http.defaults.headers.common.Authorization, 'Bearer test-token')
  assert.deepEqual(requests[0], { url: '/api/login', data: { username: 'alice', password: 'secret' } })
  await auth.logout()
  assert.equal(http.defaults.headers.common.Authorization, undefined)
  assert.equal(requests[1].url, '/api/logout')
}

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

void testAuthClient()
  .then(() => console.log('SDK contract tests passed'))
  .catch((error) => {
    console.error(error)
    process.exitCode = 1
  })
