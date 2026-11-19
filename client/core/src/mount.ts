// The non-React convenience: the stream with the keyboard/mouse overlay on top fill `el`,
// letterboxed to the remote screen. Returns what takes them out again. A stream without input is
// just `client.transport.attach(el)`; a UI framework does these two attaches itself (../src/features/video/Player.tsx).
import { Overlay } from './overlay.ts'
import type { NekoClient } from './client.ts'

export function mount(client: NekoClient, el: HTMLElement): () => void {
  const box = document.createElement('div')
  box.style.position = 'relative'
  const detach = client.transport.attach(box)
  el.append(box)
  const detachOverlay = new Overlay(client).attach(box)
  return () => {
    detachOverlay()
    detach()
    box.remove()
  }
}
