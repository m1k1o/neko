import { Members } from '@/features/members'
import { RoomMenu } from '@/features/room-menu'
import { Controls } from '@/features/controls'
import { Emotes } from '@/features/emotes'

// the bar under the video: the members, then the room menu, the controls and the emotes in three
// columns (a grid, so the columns stay when the emotes are hidden); not shown on a phone
export function RoomBar() {
  return (
    <div
      className="phone:hidden flex h-controls max-w-full shrink-0 flex-col bg-background-tertiary"
      data-testid="room-bar"
    >
      <Members />
      <div className="grid max-w-full flex-1 grid-cols-3">
        <RoomMenu />
        <Controls />
        <Emotes />
      </div>
    </div>
  )
}
