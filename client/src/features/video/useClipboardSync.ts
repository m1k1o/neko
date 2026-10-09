import { useEffect } from 'react'
import { client } from '@/state/client'

// Firefox reports readText but hangs; Safari needs a gesture per read -> use the textarea fallback there
const ua = navigator.userAgent
export const canReadClipboard =
  typeof navigator.clipboard?.readText === 'function' &&
  !ua.includes('Firefox') &&
  !(ua.includes('Safari') && !ua.includes('Chrome') && !ua.includes('Chromium'))

async function syncClipboard() {
  if (!canReadClipboard || !client.controlling || !document.hasFocus()) return
  try {
    const text = await navigator.clipboard.readText()
    if (text !== client.state.control.clipboard?.text) client.send('clipboard/set', { text })
  } catch {}
}

// the local clipboard goes to the remote when the window gets the focus; returns the sync for
// the pointer entering the video
export function useClipboardSync() {
  useEffect(() => {
    window.addEventListener('focus', syncClipboard)
    return () => window.removeEventListener('focus', syncClipboard)
  }, [])
  return syncClipboard
}
