// What each connection method builds, and how a browser that lacks one is handled. The mapping is
// the contract behind the picker: 'auto' must fall back, a forced mode must not, and no choice may
// leave a browser with a transport it cannot run.
import { test, afterEach } from 'vitest'
import assert from 'node:assert/strict'
import { transportsFor } from './neko'
import { WebCodecsTransport, WebRTCTransport } from '@m1k1o/neko'

const g = globalThis as any
const saved = { pc: g.RTCPeerConnection, vd: g.VideoDecoder, ad: g.AudioDecoder }
afterEach(() => {
  g.RTCPeerConnection = saved.pc
  g.VideoDecoder = saved.vd
  g.AudioDecoder = saved.ad
})

// which browser capabilities exist for a test
const browser = ({ webrtc, webcodecs }: { webrtc: boolean; webcodecs: boolean }) => {
  g.RTCPeerConnection = webrtc ? class {} : undefined
  g.VideoDecoder = webcodecs ? class {} : undefined
  g.AudioDecoder = webcodecs ? class {} : undefined
}

const kinds = (mode: Parameters<typeof transportsFor>[0]) => {
  const { transport, fallbackTransport } = transportsFor(mode)
  const first = typeof transport === 'function' ? transport() : transport!
  return { first: first.kind, fallback: fallbackTransport?.().kind ?? null }
}

test('auto tries WebRTC first and falls back to the websocket', () => {
  browser({ webrtc: true, webcodecs: true })
  assert.deepEqual(kinds('auto'), { first: 'webrtc', fallback: 'webcodecs-ws' })
})

test('a forced mode uses only that transport, with no fallback', () => {
  browser({ webrtc: true, webcodecs: true })
  assert.deepEqual(kinds('webrtc'), { first: 'webrtc', fallback: null })
  assert.deepEqual(kinds('websocket'), { first: 'webcodecs-ws', fallback: null })
})

test('a browser without WebRTC streams over the websocket, whatever was picked', () => {
  browser({ webrtc: false, webcodecs: true })
  for (const mode of ['auto', 'webrtc', 'websocket'] as const) {
    assert.deepEqual(kinds(mode), { first: 'webcodecs-ws', fallback: null }, mode)
  }
})

test('a browser without WebCodecs uses WebRTC, even when websocket was forced', () => {
  browser({ webrtc: true, webcodecs: false })
  for (const mode of ['auto', 'webrtc', 'websocket'] as const) {
    assert.deepEqual(kinds(mode), { first: 'webrtc', fallback: null }, mode)
  }
})

// in auto the first WebRTC gives up quickly so the switch is not a minute-long wait
test('auto builds a WebRTC transport that fails fast', () => {
  browser({ webrtc: true, webcodecs: true })
  const { transport } = transportsFor('auto')
  const auto = (typeof transport === 'function' ? transport() : transport) as WebRTCTransport
  const forced = (transportsFor('webrtc').transport as () => WebRTCTransport)()
  assert.ok(
    (auto as any).maxFailures < (forced as any).maxFailures,
    'auto must give up sooner than a forced WebRTC, which has nothing to fall back to',
  )
  assert.ok(forced instanceof WebRTCTransport && !(forced instanceof WebCodecsTransport))
})
