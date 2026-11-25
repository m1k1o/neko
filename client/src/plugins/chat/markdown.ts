// The chat's markdown: react-markdown (CommonMark; GFM strikethrough, tables and bare links;
// a newline is a line break) with the chat's own dialect on top as rehype plugins (`:name:` emoji,
// `||spoiler||`, `__underline__`), and the rules for text that comes from other users: links are
// http(s)/mailto only and anything else stays the text it was typed as, so do images unless the
// viewer turned them on (chat_images); raw HTML is text too (react-markdown's default). Markdown.tsx
// renders the tree; nothing is an HTML string.
import type { Element, ElementContent, Properties, Root } from 'hast'
import type { Plugin, PluggableList } from 'unified'
import remarkBreaks from 'remark-breaks'
import remarkGfm from 'remark-gfm'
import { emoji } from './emoji'

// Above FORMAT_LIMIT characters a message is shown as plain text; so is one the renderer fails on
// (Markdown.tsx). The limit is the server's default chat.max_length, so only a server configured
// for longer messages ever hits it, and a hostile message stays cheap (nested quotes cost micromark
// about depth²: 2000 bare `>` took 130 ms, 5000 `> ` 420 ms)
export const FORMAT_LIMIT = 512

// only these schemes become links; everything else stays text (javascript:, data:, relative, ...)
export function safeUrl(href: string): string | null {
  try {
    const u = new URL(href)
    return ['http:', 'https:', 'mailto:'].includes(u.protocol) ? u.href : null
  } catch {
    return null
  }
}

// an image is loaded from http(s) only, and only for a viewer who asked for it (chat_images): the
// browser fetching it tells its host the viewer's address
export function imageUrl(src: string): string | null {
  const url = safeUrl(src)
  return url && !url.startsWith('mailto:') ? url : null
}

// a link to an image file, by the extension of its path (https://x.io/cat.png?size=2): it gets a
// preview below it when images are on
const IMAGE_PATH = /\.(?:apng|avif|bmp|gif|jpe?g|png|svg|webp)$/i
export function isImageLink(href: string): boolean {
  const url = imageUrl(href)
  return !!url && IMAGE_PATH.test(new URL(url).pathname)
}

// f sees every element of the tree but code (its text is verbatim), parents first; it may replace
// the children, or return false to skip them
function walk(node: Root | Element, f: (el: Element) => void | false) {
  if (node.type === 'element' && f(node) === false) return
  for (const c of node.children) if (c.type === 'element' && c.tagName !== 'code') walk(c, f)
}

// the text children of el split on re: each match becomes make(match), or stays text when that is null
function splitText(el: Element, re: RegExp, make: (m: RegExpExecArray) => ElementContent | null) {
  el.children = el.children.flatMap((c) => {
    if (c.type !== 'text') return [c]
    const out: ElementContent[] = []
    let last = 0
    for (const m of c.value.matchAll(re)) {
      const node = make(m)
      if (!node) continue
      if (m.index > last) out.push({ type: 'text', value: c.value.slice(last, m.index) })
      out.push(node)
      last = m.index + m[0].length
    }
    if (last < c.value.length) out.push({ type: 'text', value: c.value.slice(last) })
    return out
  })
}

// a link to anything but http(s)/mailto ([x](javascript:...), <data:...>, a relative path) and an
// image (with images on: an image not from http(s)) stay the text they were typed as; a generated
// one (a GFM footnote's) keeps its text only
const rehypeSafe: Plugin<[{ images?: boolean }?], Root> = (options) => (tree, file) => {
  const src = String(file)
  walk(tree, (el) => {
    el.children = el.children.flatMap((c) => {
      if (c.type !== 'element') return [c]
      const unsafe =
        (c.tagName === 'img' && !(options?.images && imageUrl(String(c.properties.src ?? '')))) ||
        (c.tagName === 'a' && !safeUrl(String(c.properties.href ?? '')))
      if (!unsafe) return [c]
      return c.position
        ? [{ type: 'text', value: src.slice(c.position.start.offset, c.position.end.offset) }]
        : c.children
    })
  })
}

// __underline__: CommonMark reads it as strong; the marker in the source tells them apart
const rehypeUnderline: Plugin<[], Root> = () => (tree, file) => {
  const src = String(file)
  walk(tree, (el) => {
    if (el.tagName === 'strong' && src[el.position?.start.offset ?? -1] === '_') el.tagName = 'u'
  })
}

// ||spoiler||: a span Markdown.tsx renders click-to-reveal. Plain text only: a spoiler with
// formatting inside (||**x**||) is split by the parser first and stays as typed
const rehypeSpoiler: Plugin<[], Root> = () => (tree) => {
  walk(tree, (el) =>
    splitText(el, /\|\|(.+?)\|\|/g, (m) => ({
      type: 'element',
      tagName: 'span',
      properties: { dataSpoiler: true },
      children: [{ type: 'text', value: m[1] }],
    })),
  )
}

// :name: of a name the emoji data knows (emoji.ts): an empty span with the name, which Markdown.tsx
// renders as the emoji (Emoji.tsx: the character, or the image of a custom one)
const rehypeEmoji: Plugin<[], Root> = () => (tree) => {
  walk(tree, (el) =>
    splitText(el, /:([^:\s]+):/g, (m) =>
      emoji.names.has(m[1])
        ? {
            type: 'element',
            tagName: 'span',
            properties: { dataEmoji: m[1], title: m[0] },
            children: [],
          }
        : null,
    ),
  )
}

// with images on (after rehypeSafe, so every image left is from http(s)): ![alt](url) not inside a
// link links to itself, so that a click opens it in full, and a link to an image file gets a
// preview after it, a second link (data-preview) around the image. Markdown.tsx renders both
const rehypeImages: Plugin<[], Root> = () => (tree) => {
  const element = (tagName: string, properties: Properties, children: ElementContent[] = []): Element => ({
    type: 'element',
    tagName,
    properties,
    children,
  })
  const hasImage = (el: Element): boolean =>
    el.children.some((c) => c.type === 'element' && (c.tagName === 'img' || hasImage(c)))
  walk(tree, (el) => {
    if (el.tagName === 'a') return false // an image in a link already opens that link
    el.children = el.children.flatMap((c): ElementContent[] => {
      if (c.type !== 'element') return [c]
      if (c.tagName === 'img') return [element('a', { href: c.properties.src }, [c])]
      const href = String(c.properties.href ?? '')
      if (c.tagName !== 'a' || hasImage(c) || !isImageLink(href)) return [c]
      const preview = element('img', { src: href, alt: '', dataPreview: true })
      return [c, element('a', { href, dataPreview: true }, [preview])]
    })
  })
}

// react-markdown's options for a message (Markdown.tsx adds the components and urlTransform:
// safeUrl); images are the viewer's choice (chat_images)
export const remarkPlugins: PluggableList = [[remarkGfm, { singleTilde: false }], remarkBreaks]
const rehypeText: PluggableList = [rehypeUnderline, rehypeSpoiler, rehypeEmoji]
const rehypeNoImages: PluggableList = [rehypeSafe, ...rehypeText]
const rehypeWithImages: PluggableList = [[rehypeSafe, { images: true }], rehypeImages, ...rehypeText]
export const rehypePlugins = (images: boolean) => (images ? rehypeWithImages : rehypeNoImages)
// without images the plugins above leave none in the tree; should one get through, it is dropped here
export const disallowedElements = (images: boolean) => (images ? [] : ['img'])
