import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useNeko } from '@/state/hooks'
import { openMenu } from '@/state/actions'
import { isMuted } from '@/state/client'
import { sendChat } from '@/state/chat'
import { loadEmoji } from '@/state/emoji'
import { a11y } from '@/components/a11y'
import { Avatar } from '@/components/Avatar'
import { t } from '@/i18n'
import { Markdown } from './Markdown'
import { EmojiPicker } from './EmojiPicker'
import './emoji-sprites.scss'
import './chat.scss'

const time = (d: Date) =>
  d.toDateString() === new Date().toDateString()
    ? `Today at ${d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}`
    : d.toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })

const MAX_MESSAGE = 512

export function Chat() {
  const { app, state } = useNeko()
  const [text, setText] = useState('')
  const [picker, setPicker] = useState(false)
  const history = useRef<HTMLUListElement>(null)
  const input = useRef<HTMLTextAreaElement>(null)

  useEffect(loadEmoji, [])
  const last = app.chat[app.chat.length - 1]?.seq
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

  const lines = app.chat
  const nameOf = (m: (typeof lines)[number]) => state.sessions[m.id]?.profile.name ?? m.name
  return (
    <div className="chat">
      <ul className="chat-history" ref={history}>
        {lines.map((m, i) =>
          m.type === 'text' ? (
            <li
              key={m.seq}
              className={`message${i > 0 && lines[i - 1].id === m.id && lines[i - 1].type === 'text' ? ' bulk' : ''}`}
            >
              <div className="author" onContextMenu={(e) => openMenu(e, m.id)}>
                <Avatar seed={nameOf(m)} avatar={state.sessions[m.id]?.profile.avatar} size={40} />
              </div>
              <div className="content">
                <div className="content-head">
                  <span>{nameOf(m)}</span>
                  <span className="timestamp">{time(m.created)}</span>
                </div>
                <Markdown source={m.content} />
              </div>
            </li>
          ) : (
            <li key={m.seq} className="event">
              <div className="content" title={time(m.created)}>
                <strong>{m.name}</strong> {m.content}
              </div>
            </li>
          ),
        )}
      </ul>
      {app.chatEnabled && !isMuted() && (
        <div className="chat-send">
          <div className="accent" />
          <div className="text-container">
            <textarea
              ref={input}
              placeholder={t('send_a_message')}
              maxLength={MAX_MESSAGE}
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={onKeyDown}
            />
            {picker && <EmojiPicker onPick={onEmoji} onClose={() => setPicker(false)} />}
            <i
              className="emoji-menu fas fa-laugh"
              {...a11y('Emoji')}
              onClick={(e) => (e.stopPropagation(), setPicker(!picker))}
            />
          </div>
        </div>
      )}
    </div>
  )
}
