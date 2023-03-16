import md, { SingleNodeParserRule, HtmlOutputRule, defaultRules, State, Rules } from 'simple-markdown'
import { Component, Vue, Prop } from 'vue-property-decorator'

const { blockQuote, inlineCode, codeBlock, autolink, newline, escape, strong, text, link, url, em, u, br } =
  defaultRules

type Rule = SingleNodeParserRule & HtmlOutputRule

interface MarkdownRules extends Rules<HtmlOutputRule> {
  inlineCode: Rule
  newline: Rule
  escape: Rule
  strong: Rule
  em: Rule
  u: Rule
  blockQuote: Rule
  codeBlock: Rule
  autolink: Rule
  url: Rule
  strike: Rule
  text: Rule
  br: Rule
  emoticon: Rule
  spoiler: Rule
  user: Rule
  channel: Rule
  role: Rule
  emoji: Rule
  everyone: Rule
  here: Rule
  link?: Rule
}

interface HTMLAttributes {
  [key: string]: string
}

interface MarkdownState extends State {}

function htmlTag(
  tagName: string,
  content: string,
  attributes: HTMLAttributes,
  state: State = {},
  isClosed: boolean = true,
) {
  if (!attributes) {
    attributes = {}
  }

  if (attributes.class && state.cssModuleNames) {
    attributes.class = attributes.class
      .split(' ')
      .map((cl) => state.cssModuleNames[cl] || cl)
      .join(' ')
  }

  let attributeString = ''
  for (const attr in attributes) {
    if (Object.prototype.hasOwnProperty.call(attributes, attr) && attributes[attr]) {
      attributeString += ` ${attr}="${attributes[attr]}"` // md.sanitizeText(attr)
    }
  }

  const unclosedTag = `<${tagName}${attributeString}>`
  if (isClosed) {
    return `${unclosedTag}${content}</${tagName}>`
  }

  return unclosedTag
}

// @ts-ignore
const rules: MarkdownRules = {
  inlineCode,
  newline,
  escape,
  strong,
  em,
  u,
  link,
  codeBlock: {
    ...codeBlock,
    match: md.inlineRegex(/^```(([a-z0-9-]+?)\n+)?\n*([^]+?)\n*```/i),
    parse(capture, parse, state) {
      return {
        lang: (capture[2] || '').trim(),
        content: capture[3] || '',
        inQuote: state.inQuote || false,
      }
    },
    html(node, output, state) {
      return htmlTag('pre', htmlTag('code', md.sanitizeText(node.content), {}, state), {}, state)
    },
  },
  blockQuote: {
    ...blockQuote,
    match(source, state, prevSource) {
      return !/^$|\n *$/.test(prevSource) || state.inQuote
        ? null
        : /^( *>>> ([\s\S]*))|^( *> [^\n]+(\n *> [^\n]+)*\n?)/.exec(source)
    },
    parse(capture, parse, state) {
      const all = capture[0]
      const isBlock = Boolean(/^ *>>> ?/.exec(all))
      const removeSyntaxRegex = isBlock ? /^ *>>> ?/ : /^ *> ?/gm
      const content = all.replace(removeSyntaxRegex, '')

      state.inQuote = true
      if (!isBlock) {
        state.inline = true
      }

      const parsed = parse(content, state)

      state.inQuote = state.inQuote || false
      state.inline = state.inline || false

      return {
        content: parsed,
        type: 'blockQuote',
      }
    },
  },
  autolink: {
    ...autolink,
    parse(capture) {
      return {
        content: [
          {
            type: 'text',
            content: capture[1],
          },
        ],
        target: capture[1],
      }
    },
    html(node, output, state) {
      return htmlTag(
        'a',
        output(node.content, state),
        { href: md.sanitizeUrl(node.target) as string, target: '_blank' },
        state,
      )
    },
  },
  url: {
    ...url,
    parse(capture) {
      return {
        content: [
          {
            type: 'text',
            content: capture[1],
          },
        ],
        target: capture[1],
      }
    },
    html(node, output, state) {
      return htmlTag(
        'a',
        output(node.content, state),
        { href: md.sanitizeUrl(node.target) as string, target: '_blank' },
        state,
      )
    },
  },
  strike: {
    order: md.defaultRules.text.order,
    match: md.inlineRegex(/^~~([\s\S]+?)~~(?!_)/),
    parse(capture) {
      return {
        content: [
          {
            type: 'text',
            content: capture[1],
          },
        ],
        target: capture[1],
      }
    },
    html(node, output, state) {
      return htmlTag('s', output(node.content, state), {}, state)
    },
  },
  text: {
    ...text,
    match: (source) => /^[\s\S]+?(?=[^0-9A-Za-z\s\u00c0-\uffff-]|\n\n|\n|\w+:\S|$)/.exec(source),
    html(node, output, state) {
      if (state.escapeHTML) {
        return md.sanitizeText(node.content)
      }

      return node.content
    },
  },
  br: {
    ...br,
    match: md.anyScopeRegex(/^\n/),
  },
  emoji: {
    order: md.defaultRules.strong.order,
    match: (source) => /^:([^:\s]+):/.exec(source),
    parse(capture) {
      return {
        id: capture[1],
      }
    },
    html(node, output, state) {
      return htmlTag(
        'span',
        '',
        {
          class: `emoji`,
          'data-emoji': node.id,
          'v-tooltip.top-center': `{ content:':${node.id}:', offset: 2, delay: { show: 1000, hide: 100 } }`,
        },
        state,
      )
    },
  },
  emoticon: {
    order: md.defaultRules.text.order,
    match: (source) => /^(¯\\_\(ツ\)_\/¯)/.exec(source),
    parse(capture) {
      return {
        type: 'text',
        content: capture[1],
      }
    },
    html(node, output, state) {
      return output(node.content, state)
    },
  },
  spoiler: {
    order: 0,
    match: (source) => /^\|\|([\s\S]+?)\|\|/.exec(source),
    parse(capture, parse, state) {
      return {
        content: parse(capture[1], state),
      }
    },
    html(node, output, state) {
      return htmlTag('span', htmlTag('span', output(node.content, state), {}, state), { class: 'spoiler' }, state)
    },
  },
}

const parser = md.parserFor(rules)
const htmlOutput = md.outputFor<HtmlOutputRule, 'html'>(rules, 'html')

@Component({
  name: 'neko-markdown',
})
export default class extends Vue {
  @Prop({ required: true })
  source!: string

  render(h: any) {
    const state: MarkdownState = {
      inline: true,
      inQuote: false,
      escapeHTML: true,
      cssModuleNames: null,
    }
    return h({ template: `<div>${htmlOutput(parser(this.source, state), state)}</div>` })
  }
}
