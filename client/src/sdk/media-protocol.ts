export const MEDIA_OPCODE = {
  MOVE: 0x01,
  SCROLL: 0x02,
  KEY_DOWN: 0x03,
  KEY_UP: 0x04,
  BUTTON_DOWN: 0x05,
  BUTTON_UP: 0x06,
} as const

export type MediaInputEvent = 'mousemove' | 'wheel' | 'mousedown' | 'mouseup' | 'keydown' | 'keyup'

export type MediaInput =
  | { event: 'mousemove'; x: number; y: number }
  | { event: 'wheel'; x: number; y: number; controlKey?: boolean }
  | { event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup'; key: number }

const HEADER_BYTES = 3

function assertInteger(name: string, value: number, min: number, max: number) {
  if (!Number.isInteger(value) || value < min || value > max) {
    throw new RangeError(`${name} must be an integer between ${min} and ${max}`)
  }
}

function createPacket(opcode: number, length: number) {
  const buffer = new ArrayBuffer(HEADER_BYTES + length)
  const view = new DataView(buffer)
  view.setUint8(0, opcode)
  // The v3 server protocol uses network byte order for the header and body.
  view.setUint16(1, length, false)
  return { buffer, view }
}

/** Encode one input event using the canonical v3 WebRTC data-channel format. */
export function encodeMediaInput(input: MediaInput): ArrayBuffer {
  switch (input.event) {
    case 'mousemove': {
      assertInteger('x', input.x, 0, 0xffff)
      assertInteger('y', input.y, 0, 0xffff)
      const { buffer, view } = createPacket(MEDIA_OPCODE.MOVE, 4)
      view.setUint16(3, input.x, false)
      view.setUint16(5, input.y, false)
      return buffer
    }
    case 'wheel': {
      assertInteger('x', input.x, -0x8000, 0x7fff)
      assertInteger('y', input.y, -0x8000, 0x7fff)
      const { buffer, view } = createPacket(MEDIA_OPCODE.SCROLL, 5)
      view.setInt16(3, input.x, false)
      view.setInt16(5, input.y, false)
      view.setUint8(7, input.controlKey ? 1 : 0)
      return buffer
    }
    case 'keydown':
    case 'keyup': {
      assertInteger('key', input.key, 0, 0xffffffff)
      const opcode = input.event === 'keydown' ? MEDIA_OPCODE.KEY_DOWN : MEDIA_OPCODE.KEY_UP
      const { buffer, view } = createPacket(opcode, 4)
      view.setUint32(3, input.key, false)
      return buffer
    }
    case 'mousedown':
    case 'mouseup': {
      assertInteger('key', input.key, 0, 0xffffffff)
      const opcode = input.event === 'mousedown' ? MEDIA_OPCODE.BUTTON_DOWN : MEDIA_OPCODE.BUTTON_UP
      const { buffer, view } = createPacket(opcode, 4)
      view.setUint32(3, input.key, false)
      return buffer
    }
  }
}
