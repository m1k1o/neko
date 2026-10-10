import { Fragment, useMemo, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling } from '@m1k1o/neko'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { cn } from '@/lib/utils'
import { a11y } from '@/components/a11y'
import { chat } from './store'
import { openInApp } from './actions'
import { emoji } from './emoji'
import { parseSafe, type Node as MdNode } from './markdown'

// the legacy 0.875rem / 1.125rem at a 14px root
const code =
  'rounded-[3px] bg-background-secondary px-[3px] indent-0 font-mono text-[12.25px] leading-[15.75px] whitespace-pre-wrap'

function Spoiler({ children }: { children: React.ReactNode }) {
  const [shown, setShown] = useState(false)
  return (
    <span
      className={cn(
        'rounded px-0.5',
        shown
          ? 'cursor-default bg-background-secondary [&>span]:opacity-100'
          : 'cursor-pointer bg-background-tertiary [&>span]:opacity-0',
      )}
      data-testid="spoiler"
      data-shown={shown || undefined}
      {...(shown ? {} : a11y('Spoiler'))}
      onClick={() => setShown(true)}
    >
      <span>{children}</span>
    </span>
  )
}

// a chat message: the parser's nodes as React elements (no HTML strings)
export function Markdown({ source }: { source: string }) {
  const { canOpenInApp, linksInApp } = useStore(
    app,
    useShallow((s) => ({ canOpenInApp: s.openInApp, linksInApp: s.settings.links_in_app })),
  )
  const hosting = useStore(client.store, selectControlling)
  const emojiReady = useStore(chat, (s) => s.emojiReady) // re-render once emoji names are known
  const nodes = useMemo(() => parseSafe(source), [source])
  // open-in-app needs the plugin and control of the desktop
  const inApp = canOpenInApp && hosting
  const open = (href: string) => (e: React.MouseEvent) => {
    if (!inApp || !linksInApp) return
    e.preventDefault()
    openInApp(href)
  }

  const render = (list: MdNode[]): React.ReactNode[] =>
    list.map((n, i) => {
      switch (n.t) {
        case 'text':
          return n.v
        case 'br':
          return <br key={i} />
        case 'code':
          return (
            <code key={i} className={code}>
              {n.v}
            </code>
          )
        case 'pre':
          return (
            <pre
              key={i}
              className="my-1 block flex-1 rounded border border-background-tertiary bg-background-secondary px-1.5 py-2 text-interactive-normal"
            >
              <code className={cn(code, 'block')}>{n.v}</code>
            </pre>
          )
        case 'emoji':
          return emojiReady && emoji.names.has(n.v) ? (
            <span key={i} className="emoji" data-emoji={n.v} title={`:${n.v}:`} />
          ) : (
            `:${n.v}:`
          )
        case 'link':
          return (
            <Fragment key={i}>
              <a
                className="text-text-link underline"
                href={n.href}
                target="_blank"
                rel="noopener noreferrer"
                onClick={open(n.href)}
              >
                {render(n.c)}
              </a>
              {inApp && (
                <i
                  className="fas fa-arrow-up-right-from-square ml-[0.3em] cursor-pointer"
                  {...a11y('Open in app')}
                  onClick={() => openInApp(n.href)}
                />
              )}
            </Fragment>
          )
        case 'spoiler':
          return <Spoiler key={i}>{render(n.c)}</Spoiler>
        case 'quote':
          return (
            <blockquote key={i} className="border-l-[3px] border-background-accent pl-[3px]">
              {render(n.c)}
            </blockquote>
          )
        default: {
          const Tag = n.t // strong, em, u, s
          return (
            <Tag key={i} className={n.t === 'strong' ? 'font-extrabold' : n.t === 'em' ? 'italic' : undefined}>
              {render(n.c)}
            </Tag>
          )
        }
      }
    })

  return (
    <div className="leading-[22px] text-text-normal wrap-break-word" data-testid="chat-body">
      {render(nodes)}
    </div>
  )
}
