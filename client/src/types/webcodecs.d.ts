interface VideoDecoderConfig {
  codec: string
  hardwareAcceleration?: 'no-preference' | 'prefer-hardware' | 'prefer-software'
  optimizeForLatency?: boolean
}

interface AudioDecoderConfig {
  codec: string
  sampleRate: number
  numberOfChannels: number
}

interface EncodedVideoChunkInit {
  type: 'key' | 'delta'
  timestamp: number
  duration?: number
  data: BufferSource
  transfer?: ArrayBuffer[]
}

interface EncodedAudioChunkInit {
  type: 'key' | 'delta'
  timestamp: number
  duration?: number
  data: BufferSource
  transfer?: ArrayBuffer[]
}

declare class EncodedVideoChunk {
  constructor(init: EncodedVideoChunkInit)
}

declare class EncodedAudioChunk {
  constructor(init: EncodedAudioChunkInit)
}

interface VideoFrame {
  readonly displayWidth: number
  readonly displayHeight: number
  close(): void
}

interface AudioDataCopyToOptions {
  planeIndex: number
  format?: 'f32-planar'
}

interface AudioData {
  readonly numberOfChannels: number
  readonly numberOfFrames: number
  readonly sampleRate: number
  copyTo(destination: AllowSharedBufferSource, options: AudioDataCopyToOptions): void
  close(): void
}

declare class VideoDecoder {
  constructor(init: { output: (frame: VideoFrame) => void; error: (error: DOMException) => void })
  readonly decodeQueueSize: number
  static isConfigSupported(config: VideoDecoderConfig): Promise<{ supported?: boolean }>
  configure(config: VideoDecoderConfig): void
  decode(chunk: EncodedVideoChunk): void
  close(): void
}

declare class AudioDecoder {
  constructor(init: { output: (data: AudioData) => void; error: (error: DOMException) => void })
  readonly decodeQueueSize: number
  static isConfigSupported(config: AudioDecoderConfig): Promise<{ supported?: boolean }>
  configure(config: AudioDecoderConfig): void
  decode(chunk: EncodedAudioChunk): void
  close(): void
}
