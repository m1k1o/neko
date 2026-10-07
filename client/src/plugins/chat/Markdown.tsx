import { Fragment, useMemo, useState } from 'react'
import { useNeko } from '@/state/hooks'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import { chat } from './store'
import { openInApp } from './actions'
import { emoji } from './emoji'
import { parseSafe, type Node as MdNode } from './markdown'

function Spoiler({ children }: { children: React.ReactNode }) {
  const [shown, setShown] = useState(false)
  return (
    <span
      className={`spoiler${shown ? ' active' : ''}`}
      {...(shown ? {} : a11y('Spoiler'))}
      onClick={() => setShown(true)}
    >
      <span>{children}</span>
    </span>
  )
}

// a chat message: the parser's nodes as React elements (no HTML strings)
export function Markdown({ source }: { source: string }) {
  const { app } = useNeko() // re-render once emoji names are known
  const nodes = useMemo(() => parseSafe(source), [source])
  // open-in-app needs the plugin and control of the desktop
  const inApp = app.openInApp && client.controlling
  const open = (href: string) => (e: React.MouseEvent) => {
    if (!inApp || !app.settings.links_in_app) return
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
          return <code key={i}>{n.v}</code>
        case 'pre':
          return (
            <pre key={i}>
              <code>{n.v}</code>
            </pre>
          )
        case 'emoji':
          return chat.state.emojiReady && emoji.names.has(n.v) ? (
            <span key={i} className="emoji" data-emoji={n.v} title={`:${n.v}:`} />
          ) : (
            `:${n.v}:`
          )
        case 'link':
          return (
            <Fragment key={i}>
              <a href={n.href} target="_blank" rel="noopener noreferrer" onClick={open(n.href)}>
                {render(n.c)}
              </a>
              {inApp && (
                <i
                  className="open-in-app fas fa-arrow-up-right-from-square"
                  {...a11y('Open in app')}
                  onClick={() => openInApp(n.href)}
                />
              )}
            </Fragment>
          )
        case 'spoiler':
          return <Spoiler key={i}>{render(n.c)}</Spoiler>
        case 'quote':
          return <blockquote key={i}>{render(n.c)}</blockquote>
        default: {
          const Tag = n.t // strong, em, u, s
          return <Tag key={i}>{render(n.c)}</Tag>
        }
      }
    })

  return <div className="content-body">{render(nodes)}</div>
}
