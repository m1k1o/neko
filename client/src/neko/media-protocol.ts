const HEADER_SIZE = 24
const PROTOCOL_VERSION = 1
const VIDEO_TRACK = 1
const AUDIO_TRACK = 2

const CONTROL_VERSION = 1
const CONTROL_ACK = 1
const CONTROL_KEYFRAME_REQUEST = 2
const CONTROL_SIZE = 8

export interface EncodedMediaSample {
  track: 'video' | 'audio'
  timestamp: number
  duration?: number
  keyframe: boolean
  sequence: number
  data: Uint8Array
}

export function parseMediaSample(data: ArrayBuffer): EncodedMediaSample {
  if (data.byteLength < HEADER_SIZE) throw new Error('invalid media transport message')

  const view = new DataView(data)
  if (view.getUint8(0) !== PROTOCOL_VERSION) {
    throw new Error('unsupported media transport protocol version')
  }

  const trackID = view.getUint8(1)
  let track: EncodedMediaSample['track']
  if (trackID === VIDEO_TRACK) track = 'video'
  else if (trackID === AUDIO_TRACK) track = 'audio'
  else throw new Error('unknown media track')

  const rawDuration = Number(view.getBigInt64(12))
  return {
    track,
    timestamp: Number(view.getBigInt64(4)),
    duration: rawDuration >= 0 ? rawDuration : undefined,
    keyframe: (view.getUint8(2) & 1) !== 0,
    sequence: view.getUint32(20),
    data: new Uint8Array(data, HEADER_SIZE),
  }
}

function encodeControl(kind: number, sequence: number): ArrayBuffer {
  const buffer = new ArrayBuffer(CONTROL_SIZE)
  const view = new DataView(buffer)
  view.setUint8(0, CONTROL_VERSION)
  view.setUint8(1, kind)
  view.setUint32(4, sequence)
  return buffer
}

// encodeAck reports the newest painted video frame so the server can bound the
// number of frames in flight and drop stale ones instead of queuing them.
export function encodeAck(sequence: number): ArrayBuffer {
  return encodeControl(CONTROL_ACK, sequence)
}

// encodeKeyframeRequest asks the server for a keyframe. Sent after dropping a
// frame, since the decoder cannot use delta frames across the gap.
export function encodeKeyframeRequest(): ArrayBuffer {
  return encodeControl(CONTROL_KEYFRAME_REQUEST, 0)
}
