import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useActions } from '@/state/actions'
import { selectMuted } from '@/state/app'
import { EMOTES } from '@/state/emotes'
import { useClient } from '@/state/provider'
import { SmilePlus } from 'lucide-react'
import { IconButton } from '@/components/IconButton'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import './sprites.css'

// the emote bar in the room menu: recently used emotes and the picker
export function Emotes() {
  const client = useClient()
  const actions = useActions()
  const muted = useStore(client.store, selectMuted)
  const [recent, setRecent] = useState<string[]>(() => {
    try {
      return JSON.parse(localStorage.getItem('emote_recent') || '[]')
    } catch {
      return []
    }
  })
  const repeat = useRef(0)

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
            <IconButton
              label={e}
              className={`emote ${e}`}
              onClick={(ev) => ev.detail === 0 && actions.sendEmote(e)}
              onMouseDown={(ev) => (ev.preventDefault(), start(e))}
            />
          </li>
        ))}
        <li className="mx-[5px] text-[24px]">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <IconButton label="Emotes" data-testid="emotes-open">
                <SmilePlus className="size-6" />
              </IconButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="flex w-[220px] flex-wrap items-center justify-center"
              side="top"
              align="end"
              data-testid="emotes-menu"
            >
              {EMOTES.filter((e) => !recent.includes(e)).map((e) => (
                <DropdownMenuItem key={e} className="p-[5px]" aria-label={e} title={e} onSelect={() => pick(e)}>
                  <div className={`emote ${e}`} data-emote={e} />
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </li>
      </ul>
    </div>
  )
}
