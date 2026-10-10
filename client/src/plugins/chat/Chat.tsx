import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { openMenu } from '@/state/actions'
import { client, selectMuted } from '@/state/client'
import { Laugh } from 'lucide-react'
import { cn } from '@/lib/utils'
import { IconButton } from '@/components/IconButton'
import { Avatar } from '@/components/Avatar'
import { Popover, PopoverTrigger } from '@/components/ui/popover'
import { chat } from './store'
import { sendChat } from './actions'
import { loadEmoji } from './emoji'
import { Markdown } from './Markdown'
import { EmojiPicker } from './EmojiPicker'

const time = (d: Date) =>
  d.toDateString() === new Date().toDateString()
    ? `Today at ${d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}`
    : d.toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })

const MAX_MESSAGE = 512
// a line of the history: a message (an avatar and the content) or an event
const line = 'flex flex-1 flex-row flex-nowrap overflow-hidden pt-2.5 pr-[5px] pl-2.5 wrap-break-word select-text'

export function Chat() {
  const { t } = useTranslation()
  const sessions = useStore(client.store, (s) => s.sessions)
  const muted = useStore(client.store, selectMuted)
  const { lines, enabled } = useStore(
    chat,
    useShallow((s) => ({ lines: s.lines, enabled: s.enabled })),
  )
  const [text, setText] = useState('')
  const [picker, setPicker] = useState(false)
  const history = useRef<HTMLUListElement>(null)
  const input = useRef<HTMLTextAreaElement>(null)

  useEffect(loadEmoji, [])
  const last = lines[lines.length - 1]?.seq
  useLayoutEffect(() => {
    history.current!.scrollTop = history.current!.scrollHeight
  }, [last])

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key !== 'Enter' || e.shiftKey || e.nativeEvent.isComposing) return
    e.preventDefault()
    if (!text.trim()) return
    sendChat(text)
    setText('')
  }

  // insert :name: at the cursor, like the legacy picker
  const onEmoji = (name: string) => {
    const el = input.current!
    const code = `:${name}:`
    const at = el.selectionStart ?? text.length
    const next = (text.slice(0, at) + code + text.slice(el.selectionEnd ?? at)).slice(0, MAX_MESSAGE)
    setText(next)
    setPicker(false)
    requestAnimationFrame(() => {
      el.focus()
      el.selectionStart = el.selectionEnd = at + code.length
    })
  }

  const nameOf = (m: (typeof lines)[number]) => sessions[m.id]?.profile.name ?? m.name
  return (
    <div className="flex max-h-full max-w-full flex-1 flex-col overflow-x-hidden" data-testid="chat">
      <ul
        className="max-w-full flex-1 overflow-x-hidden overflow-y-scroll selection:bg-text-link [scrollbar-color:var(--color-background-tertiary)_transparent] [scrollbar-width:thin]"
        ref={history}
      >
        {lines.map((m, i) => {
          const bulk = i > 0 && lines[i - 1].id === m.id && lines[i - 1].type === 'text'
          return m.type === 'text' ? (
            <li key={m.seq} className={cn(line, 'text-[16px]', bulk ? 'pt-0' : 'pt-[15px]')} data-testid="chat-message">
              <div
                className={cn(
                  'mr-2.5 h-10 w-10 shrink-0 grow-0 overflow-hidden rounded-full bg-style-primary',
                  bulk && 'invisible h-0',
                )}
                data-testid="chat-author"
                onContextMenu={(e) => openMenu(e, m.id)}
              >
                <Avatar seed={nameOf(m)} avatar={sessions[m.id]?.profile.avatar} size={40} />
              </div>
              <div className="flex min-w-0 flex-1 flex-col wrap-break-word">
                {!bulk && (
                  <div className="mb-[3px] block w-full cursor-default">
                    <span className="inline-block font-medium text-text-normal">{nameOf(m)}</span>
                    {/* the legacy 0.7rem and 0.3rem at a 14px root */}
                    <span
                      className="ml-[4.2px] inline-block text-[9.8px] leading-3 font-medium text-text-muted first-letter:uppercase"
                      data-testid="chat-time"
                    >
                      {time(m.created)}
                    </span>
                  </div>
                )}
                <Markdown source={m.content} />
              </div>
            </li>
          ) : (
            <li key={m.seq} className={cn(line, 'cursor-default text-text-muted')} data-testid="chat-event">
              <div className="inline-block min-w-0 align-baseline leading-5 wrap-break-word" title={time(m.created)}>
                <strong className="font-semibold">{m.name}</strong> {m.content}
              </div>
            </li>
          )
        })}
      </ul>
      {enabled && !muted && (
        <div className="flex h-[90px] max-h-[90px] shrink-0 flex-col px-2.5 pb-2.5">
          <div className="mt-[5px] mb-2.5 h-px w-full bg-white/5" />
          <div className="relative flex h-full w-full flex-1 rounded-[5px] bg-white/5">
            <textarea
              ref={input}
              className="m-[5px] flex-1 resize-none border-0 bg-transparent text-text-normal caret-text-normal selection:bg-text-link placeholder:text-text-muted [scrollbar-color:var(--color-background-tertiary)_transparent] [scrollbar-width:thin]"
              data-testid="chat-input"
              placeholder={t('chat:send_a_message')}
              maxLength={MAX_MESSAGE}
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={onKeyDown}
            />
            <Popover open={picker} onOpenChange={setPicker}>
              <PopoverTrigger asChild>
                <IconButton label="Emoji" className="mt-2 mr-[5px] h-5 w-5" data-testid="emoji-open">
                  <Laugh className="size-5" />
                </IconButton>
              </PopoverTrigger>
              {picker && <EmojiPicker onPick={onEmoji} />}
            </Popover>
          </div>
        </div>
      )}
    </div>
  )
}
