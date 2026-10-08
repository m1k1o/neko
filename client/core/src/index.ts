// @m1k1o/neko: everything needed to talk to a neko server, usable from any framework.
// The GUI in ../src imports only from here.
export { NekoClient, type NekoClientOptions, type Pos } from './client.ts'
export { Overlay } from './overlay.ts'
export { Store, Emitter } from './store.ts'
export { NekoApi, ApiError } from './api.ts'
export * from './types.ts'
