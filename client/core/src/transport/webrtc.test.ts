// The WebRTC transport on its own: its element, the signalling it does over the SessionInfo, the
// state it reports, and what close()/suspend() leave behind.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import {
  tick,
  settle,
  reset,
  FakePeer,
  FakeElement,
  FakeChannel,
  OFFER,
  init,
  documentListeners,
} from '../test/browser.ts'

const { WebRTCTransport, DataChannelInput, Emitter, OP } = await import('../index.ts')
type Transport = InstanceType<typeof WebRTCTransport>
type State = Parameters<Parameters<Transport['on']>[1]>[0]

// a SessionInfo as the client builds it, with the socket recorded
const sessionInfo = () => {
  const out: { event: string; payload: unknown }[] = []
  const messages = new Emitter<Record<string, (p: any) => void>>()
  const s = {
    url: 'http://neko.test',
    token: 'tok',
    init: init(),
    send: (event: string, payload?: unknown) => out.push({ event, payload }),
    on: (event: string, cb: (p: any) => void) => messages.on(event, cb),
    lastMessage: Date.now(),
    autoplay: false,
  }
  return { s, out, events: () => out.map((m) => m.event), receive: (e: string, p?: unknown) => messages.emit(e, p) }
}
const setup = () => {
  reset()
  const t = new WebRTCTransport()
  const states: State[] = []
  const errors: Error[] = []
  const cursors: unknown[] = []
  t.on('state', (s) => states.push(s))
  t.on('error', (e) => errors.push(e))
  t.on('cursor.position', (p) => cursors.push(p))
  t.on('cursor.image', (i) => cursors.push(i))
  const container = new FakeElement('div')
  t.attach(container as any)
  const video = container.children[0]
  return { t, states, errors, cursors, container, video, status: () => states.at(-1)?.status }
}
const connected = async () => {
  const r = setup()
  const si = sessionInfo()
  await r.t.connect(si.s)
  si.receive('signal/provide', { sdp: OFFER })
  await settle()
  FakePeer.all[0].transition('connected')
  si.out.length = 0
  return { ...r, ...si }
}

test('supported: a static check for RTCPeerConnection', () => {
  assert.equal(WebRTCTransport.supported(), true)
  const g = globalThis as any
  const pc = g.RTCPeerConnection
  delete g.RTCPeerConnection
  assert.equal(WebRTCTransport.supported(), false)
  g.RTCPeerConnection = pc
})

test('attach: the <video> the client used to create, inside the container; detach removes it', () => {
  const { t, container, video, states } = setup()
  assert.equal(t.kind, 'webrtc')
  assert.equal(container.children.length, 1)
  assert.equal(video.tag, 'video')
  assert.equal(video.playsInline, true)
  assert.equal(video.style.cssText, 'position:absolute;inset:0;width:100%;height:100%;background:transparent')
  assert.equal(t.element, video)
  assert.deepEqual(states.at(-1), {
    status: 'disconnected',
    size: { width: 0, height: 0 },
    video: { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false },
  })
  t.detach()
  assert.deepEqual(container.children, [])
  assert.equal(t.element, null)
  t.attach(container as any) // and again
  assert.equal(container.children.length, 1)
  t.attach(container as any) // a second attach replaces the element
  assert.equal(container.children.length, 1)
})

test('connect: signal/* over the session, the track on the element, state with status and size', async () => {
  const { t, video, states, status } = setup()
  const { s, out, events, receive } = sessionInfo()
  await t.connect(s)
  assert.deepEqual(out, [{ event: 'signal/request', payload: { video: {}, audio: {} } }])
  assert.equal(status(), 'connecting')
  receive('signal/candidate', { candidate: 'early' })
  receive('signal/provide', { sdp: OFFER, iceservers: [{ urls: 'stun:x' }] })
  await settle()
  const pc = FakePeer.all[0]
  assert.deepEqual(pc.config, { iceServers: [{ urls: 'stun:x' }] })
  assert.deepEqual(pc.remoteDescription, { type: 'offer', sdp: OFFER })
  assert.deepEqual(pc.candidates, [{ candidate: 'early' }])
  assert.deepEqual(events(), ['signal/request', 'signal/answer'])
  pc.onicecandidate!({ candidate: { toJSON: () => ({ candidate: 'mine' }) } })
  pc.onicecandidate!({ candidate: null })
  assert.deepEqual(out.at(-1), { event: 'signal/candidate', payload: { candidate: 'mine' } })
  receive('signal/answer', { sdp: 'renegotiated' })
  assert.deepEqual(pc.remoteDescription, { type: 'answer', sdp: 'renegotiated' })

  const stream = { id: 'stream' }
  const receiver: Record<string, unknown> = { jitterBufferTarget: 100 }
  pc.ontrack!({ track: { kind: 'audio' }, streams: [stream], receiver })
  assert.equal(video.srcObject, null, 'the audio track rides on the video stream')
  pc.ontrack!({ track: { kind: 'video' }, streams: [stream], receiver })
  assert.equal(video.srcObject, stream)
  assert.equal(receiver.jitterBufferTarget, 0, 'the smallest jitter buffer')
  await settle()
  assert.equal(video.paused, true, 'no autoplay unless the session asks')

  pc.transition('connected')
  assert.equal(status(), 'connected')
  video.videoWidth = 1920
  video.videoHeight = 1080
  video.dispatch('resize')
  assert.deepEqual(states.at(-1)!.size, { width: 1920, height: 1080 })
  pc.transition('disconnected')
  assert.equal(status(), 'connecting')
})

test('playback: autoplay from the session; a refused play starts muted and unmutes on the first click', async () => {
  const { t, video, states } = setup()
  const { s, receive } = sessionInfo()
  s.autoplay = true
  await t.connect(s)
  receive('signal/provide', { sdp: OFFER })
  await settle()
  video.playError = Object.assign(new Error('blocked'), { name: 'NotAllowedError' })
  FakePeer.all[0].ontrack!({ track: { kind: 'video' }, streams: [{}], receiver: {} })
  await settle()
  assert.equal(video.paused, false)
  assert.equal(video.muted, true)
  assert.deepEqual(states.at(-1)!.video, {
    playable: false,
    playing: true,
    volume: 1,
    muted: true,
    mutedByAutoplay: true,
  })
  documentListeners.click[0]({})
  assert.equal(video.muted, false)
  assert.equal(states.at(-1)!.video.mutedByAutoplay, false)

  video.dispatch('canplaythrough')
  assert.equal(states.at(-1)!.video.playable, true)
  await t.setPlaying(false)
  assert.equal(video.paused, true)
  assert.equal(states.at(-1)!.video.playing, false)
  t.setVolume(0.5)
  t.setMuted(true)
  assert.deepEqual([states.at(-1)!.video.volume, states.at(-1)!.video.muted], [0.5, true])
  video.playError = Object.assign(new Error('nope'), { name: 'AbortError' })
  await assert.rejects(t.setPlaying(true), /nope/, 'other failures are the caller’s')
  video.playError = Object.assign(new Error('blocked'), { name: 'NotAllowedError' })
  await assert.rejects(t.setPlaying(true), /blocked/, 'already muted: nothing more to try')
})

test('no offer: "unavailable" after OFFER_TIMEOUT_MS, but only while the server keeps talking', async () => {
  const { t, status } = setup()
  const { s } = sessionInfo()
  s.lastMessage = Date.now() - 20_000 // nothing heard for a while
  await t.connect(s)
  await tick(8000)
  assert.equal(status(), 'connecting', 'a silent server is not "unavailable", its socket is dead')
  s.lastMessage = Date.now() // a heartbeat
  await tick(8000)
  assert.equal(status(), 'unavailable', 'the server is there and will not stream: the room without video')
  assert.equal(vi.getTimerCount(), 0)
})

test('a new track resumes playback that was running (reconnect), autoplay or not', async () => {
  const { t, video } = await connected()
  const pc = FakePeer.all[0]
  pc.ontrack!({ track: { kind: 'video' }, streams: [{ id: 1 }], receiver: {} })
  await settle()
  assert.equal(video.plays, 0, 'no autoplay')
  await t.setPlaying(true)
  assert.equal(video.plays, 1)
  pc.ontrack!({ track: { kind: 'video' }, streams: [{ id: 2 }], receiver: {} }) // the peer was replaced
  await settle()
  assert.equal(video.plays, 2, 'kept playing')
  await t.setPlaying(false)
  pc.ontrack!({ track: { kind: 'video' }, streams: [{ id: 3 }], receiver: {} })
  await settle()
  assert.equal(video.plays, 2, 'paused by the user: stays paused')
})

test('input: a DataChannelInput bound to the channel the peer opens; cursor frames become events', async () => {
  const { t, cursors } = await connected()
  assert.ok(t.input instanceof DataChannelInput)
  const dc = new FakeChannel()
  FakePeer.all[0].ondatachannel!({ channel: dc })
  dc.readyState = 'open'
  t.input.send(OP.MOVE, [2, 1], [2, 2])
  assert.deepEqual(dc.frames, [[1, 0, 4, 0, 1, 0, 2]])
  dc.onmessage!({ data: new Uint8Array([OP.CURSOR_POSITION, 0, 4, 0, 7, 0, 9]).buffer })
  assert.deepEqual(cursors.at(-1), { x: 7, y: 9 })
  dc.onmessage!({ data: new Uint8Array([OP.CURSOR_IMAGE, 0, 9, 0, 2, 0, 3, 0, 0, 0, 1, 0]).buffer })
  assert.deepEqual((cursors.at(-1) as { width: number }).width, 2)
  t.suspend()
  assert.equal(cursors.at(-1), null, 'the dropped peer takes the cursor with it')
  t.input.send(OP.MOVE, [2, 1], [2, 2])
  assert.equal(dc.frames.length, 1, 'nothing is sent on the old channel')
})

test('close: peer, timers, failures and the picture go; suspend keeps the picture and the count', async () => {
  const old = await connected()
  const { t, video, errors, status } = old
  const stream = { id: 'stream' }
  FakePeer.all[0].ontrack!({ track: { kind: 'video' }, streams: [stream], receiver: {} })
  FakePeer.all[0].transition('failed') // one failure, a retry is pending
  assert.equal(vi.getTimerCount(), 1)
  t.suspend()
  assert.ok(FakePeer.all[0].closed)
  assert.equal(vi.getTimerCount(), 0)
  assert.equal(status(), 'disconnected')
  assert.equal(video.srcObject, stream, 'a reconnect keeps the last picture')
  await tick(60_000)
  assert.equal(FakePeer.all.length, 1, 'nothing is requested while suspended')
  old.receive('signal/provide', { sdp: OFFER }) // the old session's messages are not its business any more
  await settle()
  assert.equal(FakePeer.all.length, 1)

  // the count carries over a suspend: the next failure waits the second step of the backoff
  const si = sessionInfo()
  await t.connect(si.s)
  si.receive('signal/provide', { sdp: OFFER })
  await settle()
  si.out.length = 0
  FakePeer.all[1].transition('failed')
  await tick(1500)
  assert.deepEqual(si.events(), [], 'second failure: 3000 ms')
  await tick(1500)
  assert.deepEqual(si.events(), ['signal/request'])

  t.close()
  assert.equal(video.srcObject, null, 'closed for good: the element is emptied')
  assert.equal(vi.getTimerCount(), 0)
  const again = sessionInfo()
  await t.connect(again.s)
  again.receive('signal/provide', { sdp: OFFER })
  await settle()
  again.out.length = 0
  FakePeer.all.at(-1)!.transition('failed')
  await tick(1500)
  assert.deepEqual(again.events(), ['signal/request'], 'close() reset the count: first failure waits 1500 ms')
  assert.deepEqual(errors, [])
})

test('errors: a codec mismatch and the give-up are reported, the transport stays usable', async () => {
  const { t, errors } = setup()
  const { s, out, receive } = sessionInfo()
  await t.connect(s)
  FakePeer.answerSdp = 'v=0\r\nm=audio 9 RTP/AVP 0\r\nm=video 0 RTP/AVP 96\r\n'
  receive('signal/provide', { sdp: OFFER.replace('VP8', 'AV1') })
  await settle()
  assert.deepEqual(
    errors.map((e) => e.message),
    ['this browser cannot play AV1 video'],
  )
  assert.equal(out.length, 1, 'no answer')
})

test('microphone: addTrack renegotiates through the session; removeTrack ignores an old sender', async () => {
  const { t, out } = await connected()
  assert.throws(() => new WebRTCTransport().addTrack({} as MediaStreamTrack), /not connected/)
  const pc = FakePeer.all[0]
  const sender = t.addTrack({ kind: 'audio' } as MediaStreamTrack)
  assert.deepEqual(pc.senders, [sender])
  await pc.onnegotiationneeded!()
  assert.deepEqual(out, [{ event: 'signal/offer', payload: { sdp: 'local-offer' } }])
  t.removeTrack({} as RTCRtpSender) // not this peer's: ignored, not thrown
  t.removeTrack(sender)
  assert.deepEqual(pc.senders, [])
})
