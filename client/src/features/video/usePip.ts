import { client } from '@/state/client'

export const canPip = typeof document.createElement('video').requestPictureInPicture === 'function'

// Picture-in-Picture of the stream's video element
export function usePip() {
  const request = () => client.video?.requestPictureInPicture().catch(() => {})
  return { canPip, request }
}
