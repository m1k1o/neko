import * as assert from 'node:assert/strict'
import { AuthClient } from '../src/sdk/auth'
import { ControlInputController } from '../src/sdk/control-input'
import { MediaSession } from '../src/sdk/media-session'
import { RoomClient } from '../src/sdk/room'
import { encodeMediaInput, MEDIA_OPCODE } from '../src/sdk/media-protocol'
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

async function testRoomClient() {
  const requests: Array<{ method: string; url: string }> = []
  const http = {
    async get<T>(url: string) {
      requests.push({ method: 'GET', url })
      return { data: { has_host: true, host_id: 'alice', epoch: 7 } as T }
    },
    async post<T>(url: string) {
      requests.push({ method: 'POST', url })
      return { data: true as T }
    },
  }
  const room = new RoomClient(http, '/api')
  assert.deepEqual(await room.controlStatus(), { has_host: true, host_id: 'alice', epoch: 7 })
  await room.giveControl('user/name')
  await room.resetControl()
  assert.deepEqual(requests, [
    { method: 'GET', url: '/api/room/control' },
    { method: 'POST', url: '/api/room/control/give/user%2Fname' },
    { method: 'POST', url: '/api/room/control/reset' },
  ])
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

const input = new ControlInputController({ scroll: 100, invertScroll: true })
assert.deepEqual(input.wheel({ deltaX: 2, deltaY: -20, deltaMode: 1, ctrlKey: true }), {
  x: -38,
  y: 100,
  controlKey: true,
})
assert.deepEqual(
  input.pointer(
    { clientX: 50, clientY: 25 },
    { left: 0, top: 0, width: 100, height: 50 },
    { width: 1000, height: 500 },
  ),
  { x: 500, y: 250 },
)

const sentPackets: ArrayBuffer[] = []
const fakeChannel = {
  readyState: 'open',
  binaryType: 'blob',
  onerror: null,
  onmessage: null,
  onclose: null,
  send(packet: ArrayBuffer) {
    sentPackets.push(packet)
  },
  close() {},
} as unknown as RTCDataChannel
const mediaSession = new MediaSession()
mediaSession.attachDataChannel(fakeChannel)
assert.equal(mediaSession.dataChannelOpen, true)
assert.equal(mediaSession.sendData('keydown', { key: 0x41, epoch: 9 }), true)
assert.equal(new DataView(sentPackets[0]).getUint32(11, false), 0x41)
mediaSession.close()
assert.equal(mediaSession.dataChannelOpen, false)

const mouseDown = new DataView(encodeMediaInput({ event: 'mousedown', key: 1 }))
assert.equal(mouseDown.getUint8(0), MEDIA_OPCODE.BUTTON_DOWN)
assert.equal(mouseDown.getUint16(1, false), 12)
assert.equal(mouseDown.getUint32(3, false), 0)
assert.equal(mouseDown.getUint32(7, false), 0)
assert.equal(mouseDown.getUint32(11, false), 1)

const mouseUp = new DataView(encodeMediaInput({ event: 'mouseup', key: 3 }))
assert.equal(mouseUp.getUint8(0), MEDIA_OPCODE.BUTTON_UP)
assert.equal(mouseUp.getUint16(1, false), 12)
assert.equal(mouseUp.getUint32(11, false), 3)

const epochInput = new DataView(encodeMediaInput({ event: 'keydown', key: 0xff, epoch: 0x100000001 }))
assert.equal(epochInput.getUint32(3, false), 1)
assert.equal(epochInput.getUint32(7, false), 1)
assert.equal(epochInput.getUint32(11, false), 0xff)
assert.throws(() => encodeMediaInput({ event: 'keydown', key: 1, epoch: -1 }), /epoch/)
assert.throws(() => encodeMediaInput({ event: 'keydown', key: 1, epoch: Number.MAX_SAFE_INTEGER + 1 }), /epoch/)

void testAuthClient()
  .then(testRoomClient)
  .then(() => console.log('SDK contract tests passed'))
  .catch((error) => {
    console.error(error)
    process.exitCode = 1
  })
