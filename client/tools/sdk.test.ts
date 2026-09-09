import * as assert from 'node:assert/strict'
import { AuthClient } from '../src/sdk/auth'
import { classifyNetworkQuality } from '../src/sdk/network-monitor'
import { validateSignalingMessage } from '../src/sdk/signaling'
import { mapPointerToScreen, normalizeScreenConfigurations } from '../src/neko/screen'

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

assert.deepEqual(
  normalizeScreenConfigurations([
    { width: 1280, height: 720, rate: 30 },
    { width: 640, height: 480, rate: 25 },
    { width: 1920, height: 1080, rate: 60 },
  ]),
  [
    { width: 1920, height: 1080, rate: 60 },
    { width: 1280, height: 720, rate: 30 },
  ],
)

assert.deepEqual(
  mapPointerToScreen(
    250,
    125,
    { left: 50, top: 25, width: 400, height: 200 },
    { width: 1920, height: 1080 },
  ),
  { x: 960, y: 540 },
)
assert.deepEqual(
  mapPointerToScreen(
    0,
    0,
    { left: 50, top: 25, width: 400, height: 200 },
    { width: 1920, height: 1080 },
  ),
  { x: 0, y: 0 },
)
assert.equal(
  mapPointerToScreen(50, 25, { left: 50, top: 25, width: 0, height: 200 }, { width: 1920, height: 1080 }),
  undefined,
)

void testAuthClient()
  .then(() => console.log('SDK contract tests passed'))
  .catch((error) => {
    console.error(error)
    process.exitCode = 1
  })
