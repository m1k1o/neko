import { useEffect, useRef, useState } from 'react'
import { useNeko } from '@/state/hooks'
import { client } from '@/state/client'
import './clipboard.scss'

// textarea fallback for browsers that cannot read the local clipboard (see Video.tsx)
export function Clipboard() {
  const { state } = useNeko()
  const remote = state.control.clipboard?.text ?? ''
  const [text, setText] = useState(remote)
  useEffect(() => setText(remote), [remote]) // follows what is copied on the remote while open
  const timer = useRef(0)
  return (
    <div className="clipboard" onClick={(e) => e.stopPropagation()}>
      <textarea
        autoFocus
        value={text}
        onFocus={(e) => e.target.select()}
        onChange={(e) => {
          setText(e.target.value)
          clearTimeout(timer.current)
          timer.current = window.setTimeout(() => client.send('clipboard/set', { text: e.target.value }), 500)
        }}
      />
    </div>
  )
}
