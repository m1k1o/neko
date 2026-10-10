import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { client } from '@/state/client'

// textarea fallback for browsers that cannot read the local clipboard (see Stage.tsx)
export function Clipboard() {
  const remote = useStore(client.store, (s) => s.control.clipboard?.text ?? '')
  const [text, setText] = useState(remote)
  useEffect(() => setText(remote), [remote]) // follows what is copied on the remote while open
  const timer = useRef(0)
  return (
    <div
      className="absolute right-2.5 bottom-2.5 block h-full max-h-[130px] w-full max-w-[330px] rounded-[3.5px] bg-background-primary p-[5px]"
      onClick={(e) => e.stopPropagation()}
    >
      <textarea
        className="h-full max-h-[120px] w-full max-w-[320px] resize-y border-0 bg-transparent text-text-normal selection:bg-text-normal"
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
