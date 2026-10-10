import { useRef } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { actions } from '@/state/actions'
import { app } from '@/state/app'
import { client, overlay } from '@/state/client'
import { a11y } from '@/components/a11y'
import { DropdownMenu, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Popover, PopoverTrigger } from '@/components/ui/popover'
import { Emote } from '@/features/emotes'
import { Resolution } from './Resolution'
import { Clipboard } from './Clipboard'
import { Player } from './Player'
import { useFullscreen } from './useFullscreen'
import { useClipboardSync, canReadClipboard } from './useClipboardSync'
import { usePip } from './usePip'

// an icon of the menus over the video
const icon = 'h-7.5 w-7.5 cursor-pointer rounded-[5px] bg-white/20 text-center text-[16px] leading-7.5 text-white/60'
const overlayClasses =
  'absolute inset-0 flex cursor-pointer items-center justify-center overflow-hidden bg-black/20 text-[120px]'

export function Stage({ hideControls, extraControls }: { hideControls: boolean; extraControls: boolean }) {
  const { t } = useTranslation()
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
            <div
              className={overlayClasses}
              data-testid="player-overlay"
              {...a11y('Play')}
              onClick={() => (client.unmute(), client.play().catch(() => {}))}
            >
              <i className="fas fa-play-circle" />
            </div>
          ) : (
            // only for the mute that autoplay forced; a user who muted on purpose keeps the desktop
            playing &&
            muted &&
            mutedByAutoplay && (
              <div
                className={overlayClasses}
                data-testid="player-overlay"
                {...a11y('Unmute')}
                onClick={() => client.unmute()}
              >
                <i className="fas fa-volume-up" />
              </div>
            )
          )}
        </div>
        {!fullscreen && !hideControls && (
          <ul className="absolute top-[15px] right-5">
            <li className="mb-2.5 last:mb-0">
              <i {...a11y('Fullscreen')} onClick={requestFullscreen} className={`fas fa-expand ${icon}`} />
            </li>
            {admin && (
              <li className="mb-2.5 last:mb-0">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <i
                      {...a11y('Screen resolution')}
                      data-testid="resolution-open"
                      className={`fas fa-desktop ${icon}`}
                    />
                  </DropdownMenuTrigger>
                  <Resolution />
                </DropdownMenu>
              </li>
            )}
            {!controlLocked && !implicit && (
              <li className={cn('mb-2.5 last:mb-0', extra)}>
                <i
                  className={cn(
                    `fas fa-computer-mouse ${icon}`,
                    hosted && !hosting && 'text-style-error/40',
                    !hosted && !hosting && 'text-text-normal/40',
                  )}
                  {...a11y(hosting ? t('controls.release') : t('controls.request'))}
                  onClick={() => playable && actions.toggleControl()}
                />
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
                    <i {...a11y('Clipboard')} data-testid="clipboard-open" className={`fas fa-clipboard ${icon}`} />
                  </PopoverTrigger>
                  <Clipboard />
                </Popover>
              </li>
            )}
            {pip.canPip && (
              <li className="mb-2.5 last:mb-0">
                <i
                  {...a11y('Picture-in-Picture')}
                  onClick={pip.request}
                  className={`fas fa-external-link-alt ${icon}`}
                />
              </li>
            )}
            {hosting && client.isTouchDevice && (
              <li
                className={cn('mb-2.5 last:mb-0', extra)}
                data-testid="keyboard-toggle"
                {...a11y('Keyboard')}
                onMouseDown={(e) => e.preventDefault()} // tapping the button must not take the focus the keyboard needs
                onClick={() => overlay.mobileKeyboardToggle()}
              >
                <i className={`fas fa-keyboard ${icon}`} />
              </li>
            )}
          </ul>
        )}
      </div>
    </div>
  )
}
