import { EncodedMediaSample, parseMediaSample } from './media-protocol'
import { MediaTransport } from './media-transport'

export class MediaWebSocket implements MediaTransport {
  private socket?: WebSocket
  private stopped = false

  constructor(
    private readonly url: string,
    private readonly onSample: (sample: EncodedMediaSample) => void,
    private readonly onError: (error: Error) => void,
  ) {}

  start() {
    if (this.socket || this.stopped) return

    const socket = new WebSocket(this.url)
    socket.binaryType = 'arraybuffer'
    socket.onmessage = ({ data }) => {
      try {
        if (!(data instanceof ArrayBuffer)) throw new Error('media websocket sent a non-binary message')
        this.onSample(parseMediaSample(data))
      } catch (error) {
        this.fail(error instanceof Error ? error : new Error(String(error)))
      }
    }
    socket.onerror = () => this.fail(new Error('media websocket failed'))
    socket.onclose = () => {
      if (!this.stopped) this.fail(new Error('media websocket closed'))
    }
    this.socket = socket
  }

  stop() {
    if (this.stopped) return
    this.stopped = true
    this.socket?.close()
    this.socket = undefined
  }

  send(data: ArrayBuffer) {
    if (this.stopped || this.socket?.readyState !== WebSocket.OPEN) return
    this.socket.send(data)
  }

  bufferedAmount() {
    return this.socket?.bufferedAmount ?? 0
  }

  private fail(error: Error) {
    if (this.stopped) return
    this.stop()
    this.onError(error)
  }
}
