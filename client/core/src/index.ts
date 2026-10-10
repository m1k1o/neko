// @m1k1o/neko: everything needed to talk to a neko server, usable from any framework.
// The GUI in ../src imports only from here.
export {
  NekoClient,
  selectSession,
  selectControlling,
  selectIsAdmin,
  type NekoClientOptions,
  type Pos,
} from './client.ts'
export { Overlay } from './overlay.ts'
export { mount } from './mount.ts'
export { Emitter } from './emitter.ts'
export { NekoApi, ApiError } from './api.ts'
export type {
  StreamTransport,
  SessionInfo,
  InputChannel,
  TransportKind,
  TransportStatus,
  TransportState,
  TransportEvents,
} from './transport.ts'
export { WebRTCTransport } from './transport/webrtc.ts'
export { DataChannelInput } from './input/datachannel.ts'
export { WebSocketInput } from './input/websocket.ts'
export * from './types.ts'
