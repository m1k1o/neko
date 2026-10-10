import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { actions } from '@/state/actions'
import { client, selectMuted } from '@/state/client'
import { EMOTES } from '@/state/emotes'
import { a11y, closeOn } from '@/components/a11y'
import { ContextMenu } from '@/components/ContextMenu'
import './sprites.css'

// the emote bar in the room menu: recently used emotes and the picker
export function Emotes() {
  const muted = useStore(client.store, selectMuted)
  const [recent, setRecent] = useState<string[]>(() => {
    try {
      return JSON.parse(localStorage.getItem('emote_recent') || '[]')
    } catch {
      return []
    }
  })
  const [picker, setPicker] = useState<{ x: number; y: number } | null>(null)
  const repeat = useRef(0)

  useEffect(() => (picker ? closeOn(() => setPicker(null)) : undefined), [picker])

  const pick = (emote: string) => {
    if (!recent.includes(emote)) {
      const next = [...recent.slice(-4), emote]
      setRecent(next)
      try {
        localStorage.setItem('emote_recent', JSON.stringify(next))
      } catch {}
    }
    actions.sendEmote(emote)
  }
  const stop = () => clearInterval(repeat.current)
  useEffect(() => stop, []) // the bar disappears when the member is muted, before any mouse-up
  // press and hold keeps sending, like the legacy client
  const start = (emote: string) => {
    actions.sendEmote(emote)
    stop()
    repeat.current = window.setInterval(() => actions.sendEmote(emote), 350)
  }

  if (muted) return null
  return (
    <div className="mr-2.5 self-center justify-self-end" onMouseLeave={stop} onMouseUp={stop}>
      <ul className="flex items-center justify-center">
        {recent.map((e) => (
          <li key={e} className="mx-[5px] text-[24px]">
            <div
              className={`emote ${e} cursor-pointer`}
              {...a11y(e)}
              onClick={(ev) => ev.detail === 0 && actions.sendEmote(e)}
              onMouseDown={(ev) => (ev.preventDefault(), start(e))}
            />
          </li>
        ))}
        <li className="mx-[5px] text-[24px]">
          <i
            className="fas fa-grin-beam cursor-pointer"
            data-testid="emotes-open"
            {...a11y('Emotes')}
            onClick={(e) => {
              e.stopPropagation()
              // opened with the keyboard: there is no pointer position, use the icon's
              const r = e.currentTarget.getBoundingClientRect()
              setPicker(e.detail === 0 ? { x: r.left, y: r.top } : { x: e.clientX, y: e.clientY })
            }}
          />
        </li>
      </ul>
      {picker && (
        <ContextMenu
          className="flex w-[220px] flex-wrap"
          data-testid="emotes-menu"
          style={{ left: Math.min(picker.x, innerWidth - 260), top: Math.max(picker.y - 260, 10) }}
        >
          {EMOTES.filter((e) => !recent.includes(e)).map((e) => (
            <li
              key={e}
              className="relative rounded-[3px] p-[5px] hover:bg-background-modifier-hover hover:text-interactive-hover focus:bg-background-modifier-hover focus:outline-0"
            >
              <div
                className={`emote ${e} cursor-pointer`}
                data-emote={e}
                {...a11y(e, 'menuitem')}
                onClick={() => pick(e)}
              />
            </li>
          ))}
        </ContextMenu>
      )}
    </div>
  )
}
