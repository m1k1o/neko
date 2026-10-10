import { memo, useState, type ReactNode } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling } from '@m1k1o/neko'
import ReactMarkdown, { type Components } from 'react-markdown'
import { useNeko } from '@/state/provider'
import { SquareArrowOutUpRight } from 'lucide-react'
import { IconButton } from '@/components/IconButton'
import { chat } from './store'
import { openInApp } from './actions'
import { Emoji } from './Emoji'
import { FORMAT_LIMIT, disallowedElements, rehypePlugins, remarkPlugins, safeUrl } from './markdown'

// the legacy 0.875rem / 1.125rem at a 14px root
const code =
  'rounded-[3px] bg-background-secondary px-[3px] indent-0 font-mono text-[12.25px] leading-[15.75px] whitespace-pre-wrap'

function Spoiler({ children }: { children: ReactNode }) {
  const [shown, setShown] = useState(false)
  if (shown)
    return (
      <span className="rounded bg-background-secondary px-0.5" data-testid="spoiler" data-shown>
        {children}
      </span>
    )
  return (
    <button
      type="button"
      className="cursor-pointer rounded bg-background-tertiary px-0.5 [&>span]:opacity-0"
      aria-label="Spoiler"
      title="Spoiler"
      data-testid="spoiler"
      onClick={() => setShown(true)}
    >
      <span>{children}</span>
    </button>
  )
}

// react-markdown's Markdown is a plain function (no hooks), called here so that a renderer error
// (pathological nesting, say) shows the text as it is instead of taking the chat down
function render(source: string, components: Components): ReactNode {
  if (source.length > FORMAT_LIMIT) return source
  try {
    return ReactMarkdown({
      children: source,
      remarkPlugins,
      rehypePlugins,
      components,
      urlTransform: safeUrl,
      disallowedElements,
      unwrapDisallowed: true,
    })
  } catch {
    return source
  }
}

// a chat message: the markdown (markdown.ts) as React elements, no HTML strings. memo: a new line
// in the chat does not parse the others again
export const Markdown = memo(function Markdown({ source }: { source: string }) {
  const neko = useNeko()
  const { client, app } = neko
  const { canOpenInApp, linksInApp } = useStore(
    app,
    useShallow((s) => ({ canOpenInApp: s.openInApp, linksInApp: s.settings.links_in_app })),
  )
  const hosting = useStore(client.store, selectControlling)
  useStore(chat(neko), (s) => s.emojiReady) // re-render once the emoji names are known (markdown.ts reads them)
  // open-in-app needs the plugin and control of the desktop
  const inApp = canOpenInApp && hosting
  const open = (href: string) => (e: React.MouseEvent) => {
    if (!inApp || !linksInApp) return
    e.preventDefault()
    openInApp(neko, href)
  }

  const components: Components = {
    a: ({ href = '', children }) => (
      <>
        <a
          className="text-text-link underline"
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          onClick={open(href)}
        >
          {children}
        </a>
        {inApp && (
          <IconButton label="Open in app" className="ml-[0.3em] align-middle" onClick={() => openInApp(neko, href)}>
            <SquareArrowOutUpRight className="size-3.5" />
          </IconButton>
        )}
      </>
    ),
    // the spans are the plugins' (markdown.ts): an emoji by name, or a spoiler
    span: ({ node, children }) =>
      node?.properties.dataEmoji === undefined ? (
        <Spoiler>{children}</Spoiler>
      ) : (
        <Emoji name={String(node.properties.dataEmoji)} />
      ),
    code: ({ children }) => <code className={code}>{children}</code>,
    pre: ({ children }) => (
      <pre className="my-1 block flex-1 rounded border border-background-tertiary bg-background-secondary px-1.5 py-2 text-interactive-normal">
        {children}
      </pre>
    ),
  }
  return (
    <div
      className="leading-[22px] text-text-normal wrap-break-word [&_blockquote]:border-l-[3px] [&_blockquote]:border-background-accent [&_blockquote]:pl-[3px] [&_em]:italic [&_:is(h1,h2,h3,h4,h5,h6)]:font-bold [&_ol]:list-decimal [&_p+p]:mt-[22px] [&_pre_code]:block [&_strong]:font-extrabold [&_:is(td,th)]:px-1 [&_ul]:list-disc [&_:is(ul,ol)]:pl-5"
      data-testid="chat-body"
    >
      {render(source, components)}
    </div>
  )
})
