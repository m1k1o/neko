import { useEffect, useRef } from 'react'
import { client } from '@/state/client'

// where the stream and the input overlay go: the client attaches its transport's element and
// the overlay here, letterboxed to the remote screen
export function Player() {
  const mount = useRef<HTMLDivElement>(null)
  useEffect(() => {
    client.mount(mount.current!)
    return () => client.unmount()
  }, [])
  return <div ref={mount} className="neko-mount" />
}
