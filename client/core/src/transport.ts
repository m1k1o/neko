// The seam between the session (NekoClient: auth, the main websocket, room state) and the media
// stream. A transport owns its media element and its own connection; it talks to the server only
// through the SessionInfo the client hands it. Modelled on guacamole-common-js, where Guacamole.Client
// runs over a Guacamole.Tunnel that may be HTTP or WebSocket.
import type { CursorImage, InitPayload } from './types.ts'

export type TransportKind = 'webrtc' | 'webcodecs-ws' | 'ws-mse' | 'hls'

export type TransportStatus =
  | 'disconnected'
  | 'connecting'
  | 'connected'
  // the server will not stream to this session (e.g. can_watch is off): the room works without video
  | 'unavailable'

export interface TransportState {
  status: TransportStatus
  // the stream's own size once a frame has arrived (letterboxing uses the server's screen size)
  size: { width: number; height: number }
  video: {
    playable: boolean
    playing: boolean
    volume: number
    muted: boolean
    // the browser refused to autoplay with sound, so playback started muted
    mutedByAutoplay: boolean
  }
}

export interface TransportEvents {
  state: (state: TransportState) => void
  // bitrate, fps, rtt, ... (nothing emits it yet)
  stats: (stats: Record<string, number>) => void
  // the stream is lost for good; the client closes the connection with it
  error: (error: Error) => void
  // the host's cursor, for transports whose server feeds it back (the WebRTC data channel);
  // null when the stream is dropped
  'cursor.position': (pos: { x: number; y: number } | null) => void
  'cursor.image': (image: CursorImage) => void
}

// what a transport gets from the client: where the server is, who we are, and the main websocket
// for signalling
export interface SessionInfo {
  url: string // http(s) origin and path of the server
  token?: string
  init: InitPayload // the system/init payload
  send(event: string, payload?: unknown): void
  on(event: string, cb: (payload: any) => void): () => void
}

// input to the remote desktop, in the server's terms (keysyms, button codes, touch ids); how it
// travels is the channel's business
export interface InputChannel {
  move(x: number, y: number): void
  scroll(deltaX: number, deltaY: number, controlKey: boolean): void
  button(code: number, down: boolean): void
  key(keysym: number, down: boolean): void
  touch(phase: 'begin' | 'update' | 'end', id: number, x: number, y: number, pressure: number): void
}

export interface StreamTransport {
  readonly kind: TransportKind
  // the media element created by attach(), for Picture-in-Picture and the like
  readonly element: HTMLElement | null
  // the transport's own input path, when it has one (the WebRTC data channel)
  readonly input?: InputChannel

  // start streaming for a session; called again on every reconnect, also while the previous
  // session's stream is still up: it is replaced (its last picture stays until the new one shows)
  connect(session: SessionInfo): Promise<void>
  // stop for good: the media path, timers, counters and the picture on the element go
  close(): void

  // create the media element inside the container; returns what removes it again
  attach(container: HTMLElement): () => void

  setPlaying(on: boolean): Promise<void>
  setVolume(volume: number): void
  setMuted(on: boolean): void

  // share local media (microphone), where the transport can; returns what stops sharing it
  shareMedia?(stream: MediaStream): () => void

  on<K extends keyof TransportEvents>(event: K, cb: TransportEvents[K]): () => void
}
