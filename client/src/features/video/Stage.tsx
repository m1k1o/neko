import { useRef } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import {
  CirclePlay,
  Clipboard as ClipboardIcon,
  Expand,
  Keyboard,
  Monitor,
  Mouse,
  PictureInPicture2,
  Volume2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import { useNeko } from '@/state/provider'
import { IconButton } from '@/components/IconButton'
import { DropdownMenu, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Popover, PopoverTrigger } from '@/components/ui/popover'
import { Emote } from '@/features/emotes/Emote'
import { Resolution } from './Resolution'
import { Clipboard } from './Clipboard'
import { Player } from './Player'
import { useFullscreen } from './useFullscreen'
import { useClipboardSync, canReadClipboard } from './useClipboardSync'
import { usePip } from './usePip'

// the play / unmute button covering the video
const overlayClasses =
  'absolute inset-0 flex cursor-pointer items-center justify-center overflow-hidden rounded-none bg-black/20'

export function Stage({ hideControls, extraControls }: { hideControls: boolean; extraControls: boolean }) {
  const { t } = useTranslation()
  const { client, overlay, app } = useNeko()
  const actions = useActions()
  const { admin, hosting, hosted, implicit, lockedControls, video } = useStore(
    client.store,
    useShallow((s) => ({
      admin: selectIsAdmin(s),
      hosting: selectControlling(s),
      hosted: s.control.host_id !== null,
      implicit: s.settings.implicit_hosting,
      lockedControls: s.settings.locked_controls,
      video: s.video,
    })),
  )
  const emotes = useStore(app, (s) => s.emotes)
  const player = useRef<HTMLDivElement>(null)
  const { fullscreen, request: requestFullscreen } = useFullscreen(player)
  const syncClipboard = useClipboardSync()
  const pip = usePip()

  const controlLocked = lockedControls && !admin
  const { playing, playable, muted, mutedByAutoplay } = video
  // usually the extra controls are only shown on a phone
  const extra = extraControls ? '' : 'phone:block hidden'

  return (
    <div className="relative h-full w-full">
      <div ref={player} className="absolute inset-0 flex items-center justify-center bg-black">
        <div className="absolute inset-0" data-testid="player" onMouseEnter={syncClipboard}>
          <Player transport={client.transport} />
          <div className="pointer-events-none absolute inset-0 overflow-hidden">
            {Object.entries(emotes).map(([id, type]) => (
              <Emote key={id} id={id} type={type} />
            ))}
          </div>
          {!playing && playable ? (
            <IconButton
              label="Play"
              className={overlayClasses}
              data-testid="player-overlay"
              onClick={() => (client.unmute(), client.play().catch(() => {}))}
            >
              <CirclePlay className="size-[120px]" />
            </IconButton>
          ) : (
            // only for the mute that autoplay forced; a user who muted on purpose keeps the desktop
            playing &&
            muted &&
            mutedByAutoplay && (
              <IconButton
                label="Unmute"
                className={overlayClasses}
                data-testid="player-overlay"
                onClick={() => client.unmute()}
              >
                <Volume2 className="size-[120px]" />
              </IconButton>
            )
          )}
        </div>
        {!fullscreen && !hideControls && (
          <ul className="absolute top-[15px] right-5">
            <li className="mb-2.5 last:mb-0">
              <IconButton variant="video" label="Fullscreen" onClick={requestFullscreen}>
                <Expand className="size-4" />
              </IconButton>
            </li>
            {admin && (
              <li className="mb-2.5 last:mb-0">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <IconButton variant="video" label="Screen resolution" data-testid="resolution-open">
                      <Monitor className="size-4" />
                    </IconButton>
                  </DropdownMenuTrigger>
                  <Resolution />
                </DropdownMenu>
              </li>
            )}
            {!controlLocked && !implicit && (
              <li className={cn('mb-2.5 last:mb-0', extra)}>
                <IconButton
                  variant="video"
                  label={hosting ? t('controls.release') : t('controls.request')}
                  className={cn(
                    hosted && !hosting && 'text-style-error/40',
                    !hosted && !hosting && 'text-text-normal/40',
                  )}
                  onClick={() => playable && actions.toggleControl()}
                >
                  <Mouse className="size-4" />
                </IconButton>
              </li>
            )}
          </ul>
        )}
        {!fullscreen && !hideControls && (
          <ul className="absolute right-5 bottom-[15px]">
            {hosting && !canReadClipboard && (
              <li className="mb-2.5 last:mb-0">
                <Popover>
                  <PopoverTrigger asChild>
                    <IconButton variant="video" label="Clipboard" data-testid="clipboard-open">
                      <ClipboardIcon className="size-4" />
                    </IconButton>
                  </PopoverTrigger>
                  <Clipboard />
                </Popover>
              </li>
            )}
            {pip.canPip && (
              <li className="mb-2.5 last:mb-0">
                <IconButton variant="video" label="Picture-in-Picture" onClick={pip.request}>
                  <PictureInPicture2 className="size-4" />
                </IconButton>
              </li>
            )}
            {hosting && client.isTouchDevice && (
              <li className={cn('mb-2.5 last:mb-0', extra)}>
                <IconButton
                  variant="video"
                  label="Keyboard"
                  data-testid="keyboard-toggle"
                  onMouseDown={(e) => e.preventDefault()} // tapping the button must not take the focus the keyboard needs
                  onClick={() => overlay.mobileKeyboardToggle()}
                >
                  <Keyboard className="size-4" />
                </IconButton>
              </li>
            )}
          </ul>
        )}
      </div>
    </div>
  )
}
