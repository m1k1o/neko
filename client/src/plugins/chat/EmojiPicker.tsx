import { useEffect, useRef, useState } from 'react'
import { useNeko } from '@/state/hooks'
import { emoji, pickedEmoji } from '@/state/emoji'
import { a11y, closeOn } from '@/components/a11y'
import './emoji.scss'

export function EmojiPicker({ onPick, onClose }: { onPick: (name: string) => void; onClose: () => void }) {
  const { app } = useNeko()
  const [search, setSearch] = useState('')
  const [hovered, setHovered] = useState('')
  const [active, setActive] = useState(0)
  const ref = useRef<HTMLDivElement>(null)
  const scroll = useRef<HTMLDivElement>(null)
  const groupEls = useRef<(HTMLLIElement | null)[]>([])

  useEffect(() => closeOn(onClose), [onClose])

  const groups = [{ id: 'recent', name: 'Recent', list: app.emojiRecent }, ...emoji.groups]
  const q = search.trim().toLowerCase()
  const filtered = q
    ? [...emoji.names].filter((n) => n.includes(q) || emoji.keywords[n]?.some((k) => k.includes(q)))
    : []

  const pick = (name: string) => {
    pickedEmoji(name)
    onPick(name)
  }
  const item = (name: string, key: string) => (
    <li key={key} className={`emoji-container${hovered === name ? ' active' : ''}`}>
      <span
        className="emoji"
        data-emoji={name}
        {...a11y(`:${name}:`)}
        tabIndex={-1}
        onMouseEnter={() => setHovered(name)}
        onFocus={() => setHovered(name)}
        onClick={() => pick(name)}
      />
    </li>
  )
  const onScroll = () => {
    const top = scroll.current!.scrollTop
    let i = 0
    groupEls.current.forEach((el, idx) => el && el.offsetTop <= top && (i = idx))
    setActive(i)
  }

  return (
    <div className="neko-emoji" ref={ref} onClick={(e) => e.stopPropagation()}>
      <div className="search">
        <div className="search-contianer">
          <input
            type="text"
            autoFocus
            value={search}
            placeholder={hovered ? `:${hovered}:` : ''}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && filtered[0] && pick(filtered[0])}
          />
        </div>
      </div>
      <div className="list" ref={scroll} onScroll={onScroll}>
        {q ? (
          <ul className="emoji-container" style={{ display: 'flex' }}>
            {filtered.map((n) => item(n, n))}
          </ul>
        ) : (
          <ul className="group-list">
            {groups.map((g, gi) => (
              <li key={g.id} className="group" ref={(el) => void (groupEls.current[gi] = el)}>
                <span className="label">{g.name}</span>
                <ul className="emoji-list">{g.list.map((n) => item(n, `${g.id}-${n}`))}</ul>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="details">
        {hovered && (
          <div className="details-container">
            <span className="emoji" data-emoji={hovered} />
            <span className="emoji-id">:{hovered}:</span>
          </div>
        )}
      </div>
      <div className="groups">
        <ul>
          {groups.map((g, gi) => (
            <li
              key={g.id}
              className={`${g.id}${active === gi && !q ? ' active' : ''}`}
              {...a11y(g.name, 'tab')}
              onClick={() => (scroll.current!.scrollTop = gi === 0 ? 0 : (groupEls.current[gi]?.offsetTop ?? 0))}
            >
              <span className={`group-${g.id} fas`} />
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
