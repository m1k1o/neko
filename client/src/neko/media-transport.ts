export interface MediaTransport {
  start(): void | Promise<void>
  stop(): void
  // send delivers a control message back to the server. It is best effort:
  // control messages are dropped while the socket is not open.
  send(data: ArrayBuffer): void
  // bufferedAmount reports bytes queued but not yet written to the network.
  // A growing value means the uplink, not the decoder, is the bottleneck.
  bufferedAmount(): number
}

export const WEB_CODECS_MODES = ['webcodecs-ws'] as const

export function isWebCodecsMode(value: string | null): boolean {
  return WEB_CODECS_MODES.some((mode) => mode === value)
}
