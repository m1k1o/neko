import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { app } from '@/state/app'
import { client, overlay } from '@/state/client'
import { a11y, closeOn } from '@/components/a11y'
import { Emote } from '@/features/emotes'
import { Resolution } from './Resolution'
import { Clipboard } from './Clipboard'
import { Player } from './Player'
import { useFullscreen } from './useFullscreen'
import { useClipboardSync, canReadClipboard } from './useClipboardSync'
import { usePip } from './usePip'
import './video.scss'

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
  const [menu, setMenu] = useState<null | 'resolution' | 'clipboard'>(null)
  const { fullscreen, request: requestFullscreen } = useFullscreen(player)
  const syncClipboard = useClipboardSync()
  const pip = usePip()

  useEffect(() => (menu ? closeOn(() => setMenu(null)) : undefined), [menu])

  const controlLocked = lockedControls && !admin
  const { playing, playable, muted, mutedByAutoplay } = video
  const open = (m: 'resolution' | 'clipboard') => (e: React.MouseEvent) => (
    e.stopPropagation(),
    setMenu(menu === m ? null : m)
  )
  const extra = extraControls ? '' : 'extra-control'

  return (
    <div className="video">
      <div ref={player} className="player">
        <div className="player-container" data-testid="player" onMouseEnter={syncClipboard}>
          <Player transport={client.transport} />
          <div className="emotes">
            {Object.entries(emotes).map(([id, type]) => (
              <Emote key={id} id={id} type={type} />
            ))}
          </div>
          {!playing && playable ? (
            <div
              className="player-overlay"
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
                className="player-overlay"
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
          <ul className="video-menu top">
            <li>
              <i {...a11y('Fullscreen')} onClick={requestFullscreen} className="fas fa-expand" />
            </li>
            {admin && (
              <li>
                <i
                  {...a11y('Screen resolution')}
                  data-testid="resolution-open"
                  onClick={open('resolution')}
                  className="fas fa-desktop"
                />
              </li>
            )}
            {!controlLocked && !implicit && (
              <li className={extra}>
                <i
                  className={`${hosted && !hosting ? 'disabled ' : ''}${!hosted && !hosting ? 'faded ' : ''}fas fa-computer-mouse`}
                  {...a11y(hosting ? t('controls.release') : t('controls.request'))}
                  onClick={() => playable && actions.toggleControl()}
                />
              </li>
            )}
          </ul>
        )}
        {!fullscreen && !hideControls && (
          <ul className="video-menu bottom">
            {hosting && !canReadClipboard && (
              <li>
                <i
                  {...a11y('Clipboard')}
                  data-testid="clipboard-open"
                  onClick={open('clipboard')}
                  className="fas fa-clipboard"
                />
              </li>
            )}
            {pip.canPip && (
              <li>
                <i {...a11y('Picture-in-Picture')} onClick={pip.request} className="fas fa-external-link-alt" />
              </li>
            )}
            {hosting && client.isTouchDevice && (
              <li
                className={extra}
                data-testid="keyboard-toggle"
                {...a11y('Keyboard')}
                onMouseDown={(e) => e.preventDefault()} // tapping the button must not take the focus the keyboard needs
                onClick={() => overlay.mobileKeyboardToggle()}
              >
                <i className="fas fa-keyboard" />
              </li>
            )}
          </ul>
        )}
        {menu === 'resolution' && admin && <Resolution onPick={() => setMenu(null)} />}
        {menu === 'clipboard' && hosting && <Clipboard />}
      </div>
    </div>
  )
}
