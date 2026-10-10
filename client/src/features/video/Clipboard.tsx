import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useClient } from '@/state/provider'
import { PopoverContent } from '@/components/ui/popover'

// textarea fallback for browsers that cannot read the local clipboard (see Stage.tsx), a popover
// left of the icon that opens it
export function Clipboard() {
  const client = useClient()
  const remote = useStore(client.store, (s) => s.control.clipboard?.text ?? '')
  const [text, setText] = useState(remote)
  useEffect(() => setText(remote), [remote]) // follows what is copied on the remote while open
  const timer = useRef(0)
  return (
    <PopoverContent
      side="left"
      align="end"
      className="block h-[130px] w-[330px] rounded-[3.5px] bg-background-primary p-[5px]"
      data-testid="clipboard"
    >
      <textarea
        className="h-full w-full resize-y border-0 bg-transparent text-text-normal selection:bg-text-normal"
        aria-label="Clipboard"
        value={text}
        onFocus={(e) => e.target.select()}
        onChange={(e) => {
          setText(e.target.value)
          clearTimeout(timer.current)
          timer.current = window.setTimeout(() => client.send('clipboard/set', { text: e.target.value }), 500)
        }}
      />
    </PopoverContent>
  )
}
