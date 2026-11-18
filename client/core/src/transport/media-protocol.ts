// The wire format of the media websocket (server/internal/media). Frames are binary: a fixed
// header then the encoded sample. Control messages travel back on the same socket.

// `transfer` is in the WebCodecs spec (and in Chromium) but missing from the DOM library at the
// TypeScript version this project pins. It hands the sample's buffer to the decoder instead of
// copying it, which is worth keeping on a path that runs for every frame.
declare global {
  interface EncodedVideoChunkInit {
    transfer?: ArrayBuffer[]
  }
  interface EncodedAudioChunkInit {
    transfer?: ArrayBuffer[]
  }
}

export const SAMPLE_HEADER_SIZE = 24
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
  // the server's frame counter, echoed back in an ack so it can bound frames in flight
  sequence: number
  data: Uint8Array
}

export function parseMediaSample(data: ArrayBuffer): EncodedMediaSample {
  if (data.byteLength < SAMPLE_HEADER_SIZE) throw new Error('invalid media transport message')

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
    data: new Uint8Array(data, SAMPLE_HEADER_SIZE),
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

// encodeAck reports the newest painted frame so the server can advance its window and drop stale
// frames instead of queuing them.
export const encodeAck = (sequence: number) => encodeControl(CONTROL_ACK, sequence)

// encodeKeyframeRequest asks for a keyframe after a dropped frame: the decoder cannot use delta
// frames across the gap, so without one the picture stays frozen until the encoder's next keyframe.
export const encodeKeyframeRequest = () => encodeControl(CONTROL_KEYFRAME_REQUEST, 0)
