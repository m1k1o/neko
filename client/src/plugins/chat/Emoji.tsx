import { emoji, custom } from './emoji'

// an emoji by name in a 22px box: its Unicode character (the browser's emoji font), or the image of
// a custom one
export function Emoji({ name }: { name: string }) {
  const title = `:${name}:`
  return name in custom ? (
    <img className="inline-block h-[22px] w-[22px] align-bottom" src={custom[name]} alt={title} data-emoji={name} />
  ) : (
    <span
      className="inline-block h-[22px] w-[22px] text-center align-bottom text-[18px] leading-[22px]"
      data-emoji={name}
      title={title}
    >
      {emoji.chars[name]}
    </span>
  )
}
