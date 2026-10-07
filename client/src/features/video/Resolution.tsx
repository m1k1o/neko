import { useNeko } from '@/state/hooks'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import './resolution.scss'

export function Resolution({ onPick }: { onPick: () => void }) {
  const { state } = useNeko()
  const { width, height, rate } = state.screen.size
  return (
    <ul className="resolution" role="menu" style={{ top: 50, right: 50 }} onClick={(e) => e.stopPropagation()}>
      {state.screen.configurations.map((c, i) => (
        <li
          key={i}
          className={c.width === width && c.height === height && c.rate === rate ? 'active' : ''}
          {...a11y(`${c.width}x${c.height}@${c.rate}`, 'menuitem')}
          onClick={() => (client.setScreenSize(c.width, c.height, c.rate), onPick())}
        >
          <i className="fas fa-desktop"></i>
          <span>
            {c.width}x{c.height}
          </span>
          <small>{c.rate}</small>
        </li>
      ))}
    </ul>
  )
}
