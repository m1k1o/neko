import { useEffect, useRef } from 'react'
import { hideEmote } from '@/state/emotes'
import './sprites.scss'
import './emote.scss'

const rnd = (a: number, b: number) => a + Math.random() * (b - a)

// one emote flying over the video: seven copies float up and fade, like the legacy anime.js version
export function Emote({ id, type }: { id: string; type: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const anims = [...ref.current!.children].map((el, i) => {
      const side = (odd: boolean) => (odd ? rnd(0, 50) : rnd(-50, 0)) + '%'
      return (el as HTMLElement).animate(
        [
          { left: side(!!(i % 2)), top: '0%', opacity: 0, transform: 'rotate(0deg)' },
          { left: side(!!(i % 2)), opacity: 1, offset: 0.33 },
          { left: side(!(i % 2)), opacity: 0.5, offset: 0.66 },
          { left: side(!!(i % 2)), top: rnd(-600, -200) + '%', opacity: 0, transform: `rotate(${rnd(-35, 35)}deg)` },
        ],
        { duration: rnd(1000, 2000), easing: 'ease-in-out', fill: 'forwards' },
      ).finished
    })
    Promise.all(anims).then(() => hideEmote(id))
  }, [id])
  return (
    <div ref={ref} className="emote-container">
      {Array.from({ length: 7 }, (_, i) => (
        <div key={i} className={`emote ${type}`} />
      ))}
    </div>
  )
}
