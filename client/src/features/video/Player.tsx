import { useEffect, useRef } from 'react'
import type { StreamTransport } from '@m1k1o/neko'
import { overlay } from '@/state/client'

// the stream, with the input overlay on top: the transport draws into the box first, then the
// overlay, which keeps the box letterboxed to the remote screen
export function Player({ transport }: { transport: StreamTransport }) {
  const box = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const detach = transport.attach(box.current!)
    const detachOverlay = overlay.attach(box.current!)
    return () => {
      detachOverlay()
      detach()
    }
  }, [transport])
  return (
    <div className="neko-mount">
      <div ref={box} style={{ position: 'relative' }} />
    </div>
  )
}
