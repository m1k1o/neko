import { useEffect, useState, type RefObject } from 'react'
import { useClient } from '@/state/provider'

// fullscreen on `target`; while fullscreen the keyboard lock lets Esc/Alt+Tab etc. reach the remote
export function useFullscreen(target: RefObject<HTMLElement | null>) {
  const client = useClient()
  const [fullscreen, setFullscreen] = useState(false)
  useEffect(() => {
    const onFs = () => {
      const fs = !!document.fullscreenElement
      setFullscreen(fs)
      const kb = (navigator as any).keyboard
      fs ? kb?.lock?.().catch(() => {}) : kb?.unlock?.()
    }
    document.addEventListener('fullscreenchange', onFs)
    return () => document.removeEventListener('fullscreenchange', onFs)
  }, [])
  const request = () => {
    // iOS only allows fullscreen on the video element itself
    if (target.current?.requestFullscreen) target.current.requestFullscreen().catch(() => {})
    else (client.transport.element as any)?.webkitEnterFullscreen?.()
  }
  return { fullscreen, request }
}
