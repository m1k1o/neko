import { useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import {
  Apple,
  Car,
  Cat,
  Clock,
  Flag,
  Lightbulb,
  PawPrint,
  Search,
  Shapes,
  Smile,
  Users,
  Volleyball,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { IconButton } from '@/components/IconButton'
import { PopoverContent } from '@/components/ui/popover'
import { useNeko } from '@/state/provider'
import { chat } from './store'
import { emoji, pickedEmoji } from './emoji'
import { Emoji } from './Emoji'

// the icon of each group's tab at the bottom of the picker
const GROUP_ICON: Record<string, LucideIcon> = {
  recent: Clock,
  neko: Cat,
  emotion: Smile,
  people: Users,
  nature: PawPrint,
  food: Apple,
  activity: Volleyball,
  travel: Car,
  objects: Lightbulb,
  symbols: Shapes,
  flags: Flag,
}

export function EmojiPicker({ onPick }: { onPick: (name: string) => void }) {
  const store = chat(useNeko())
  const { recent, ready } = useStore(
    store,
    useShallow((s) => ({ recent: s.emojiRecent, ready: s.emojiReady })),
  )
  const [search, setSearch] = useState('')
  const [hovered, setHovered] = useState('')
  const [active, setActive] = useState(0)
  const scroll = useRef<HTMLDivElement>(null)
  const groupEls = useRef<(HTMLLIElement | null)[]>([])

  const groups = [{ id: 'recent', name: 'Recent', list: recent }, ...(ready ? emoji.groups : [])]
  const q = search.trim().toLowerCase()
  const filtered = q
    ? [...emoji.names].filter((n) => n.includes(q) || emoji.keywords[n]?.some((k) => k.includes(q)))
    : []

  const pick = (name: string) => {
    pickedEmoji(store, name)
    onPick(name)
  }
  const item = (name: string, key: string) => (
    <li key={key} className={cn('cursor-pointer rounded-[3px] p-0.5', hovered === name && 'bg-background-floating')}>
      <IconButton
        label={`:${name}:`}
        tabIndex={-1}
        onMouseEnter={() => setHovered(name)}
        onFocus={() => setHovered(name)}
        onClick={() => pick(name)}
      >
        <Emoji name={name} />
      </IconButton>
    </li>
  )
  const onScroll = () => {
    const top = scroll.current!.scrollTop
    let i = 0
    groupEls.current.forEach((el, idx) => el && el.offsetTop <= top && (i = idx))
    setActive(i)
  }

  return (
    <PopoverContent
      side="top"
      align="end"
      sideOffset={8}
      className="flex h-[350px] w-[300px] flex-col overflow-hidden rounded-[5px] bg-background-secondary shadow-elevation-high"
      data-testid="emoji-picker"
    >
      <div className="shrink-0 border-b border-background-tertiary p-2.5">
        <div className="relative flex flex-col overflow-hidden rounded-[5px] text-interactive-normal">
          <Search className="absolute top-1.5 right-1.5 size-[15px] opacity-50" />
          <input
            className="border-0 bg-background-floating p-[5px] leading-4 font-medium text-interactive-normal placeholder:font-medium placeholder:text-text-muted"
            type="text"
            data-testid="emoji-search"
            value={search}
            placeholder={hovered ? `:${hovered}:` : ''}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && filtered[0] && pick(filtered[0])}
          />
        </div>
      </div>
      <div
        className="relative grow scroll-smooth overflow-x-hidden overflow-y-scroll p-[5px] [scrollbar-color:var(--color-background-tertiary)_transparent] [scrollbar-width:thin]"
        data-testid="emoji-list"
        ref={scroll}
        onScroll={onScroll}
      >
        {q ? (
          <ul className="flex">{filtered.map((n) => item(n, n))}</ul>
        ) : (
          <ul className="flex w-[300px] flex-col">
            {groups.map((g, gi) => (
              <li key={g.id} ref={(el) => void (groupEls.current[gi] = el)}>
                <span className="sticky top-[-5px] z-2 block w-full bg-background-secondary/90 py-2 text-[12px] font-medium uppercase">
                  {g.name}
                </span>
                <ul className="flex flex-row flex-wrap">{g.list.map((n) => item(n, `${g.id}-${n}`))}</ul>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="flex h-9 shrink-0 flex-col justify-center bg-background-tertiary">
        {hovered && (
          <div className="flex h-5 flex-row">
            <span className="mr-[5px] ml-2.5 cursor-default">
              <Emoji name={hovered} />
            </span>
            <span className="cursor-default text-[16px] leading-5 font-medium">:{hovered}:</span>
          </div>
        )}
      </div>
      <div className="h-7.5 shrink-0 bg-background-floating px-[5px]">
        <ul className="flex flex-row flex-wrap">
          {groups.map((g, gi) => {
            const Icon = GROUP_ICON[g.id]
            return (
              <li key={g.id} className="flex grow">
                <IconButton
                  label={g.name}
                  className={cn(
                    'box-content h-[27px] w-full',
                    active === gi && !q && 'border-b-[3px] border-style-primary',
                  )}
                  aria-current={active === gi && !q ? 'true' : undefined}
                  onClick={() => (scroll.current!.scrollTop = gi === 0 ? 0 : (groupEls.current[gi]?.offsetTop ?? 0))}
                >
                  <Icon className="size-4" />
                </IconButton>
              </li>
            )
          })}
        </ul>
      </div>
    </PopoverContent>
  )
}
