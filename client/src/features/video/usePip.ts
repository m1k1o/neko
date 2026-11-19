import { useClient } from '@/state/provider'

export const canPip = typeof document.createElement('video').requestPictureInPicture === 'function'

// Picture-in-Picture of the stream's video element
export function usePip() {
  const client = useClient()
  const request = () => {
    const el = client.transport.element
    if (el instanceof HTMLVideoElement) el.requestPictureInPicture().catch(() => {})
  }
  return { canPip, request }
}
