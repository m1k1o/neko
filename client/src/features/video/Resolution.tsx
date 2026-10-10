import { useStore } from 'zustand'
import { client } from '@/state/client'
import { a11y } from '@/components/a11y'
import './resolution.scss'

export function Resolution({ onPick }: { onPick: () => void }) {
  const { size, configurations } = useStore(client.store, (s) => s.screen)
  const { width, height, rate } = size
  return (
    <ul
      className="resolution"
      role="menu"
      data-testid="resolution-menu"
      style={{ top: 50, right: 50 }}
      onClick={(e) => e.stopPropagation()}
    >
      {configurations.map((c, i) => (
        <li
          key={i}
          className={c.width === width && c.height === height && c.rate === rate ? 'active' : ''}
          data-testid="resolution-item"
          data-active={(c.width === width && c.height === height && c.rate === rate) || undefined}
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
