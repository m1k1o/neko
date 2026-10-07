import { Members } from '@/features/members'
import { RoomMenu } from '@/features/room-menu'
import { Controls } from '@/features/controls'
import { Emotes } from '@/features/emotes'

// the bar under the video: members, room menu, controls and emotes (layout in app/app.scss)
export function RoomBar() {
  return (
    <div className="room-container">
      <Members />
      <div className="room-menu">
        <div className="settings">
          <RoomMenu />
        </div>
        <div className="controls">
          <Controls />
        </div>
        <div className="emotes">
          <Emotes />
        </div>
      </div>
    </div>
  )
}
