import { useEffect, useRef, useState } from 'react'
import { useNeko } from '@/state/hooks'
import { actions } from '@/state/actions'
import { client } from '@/state/client'
import { a11y, closeOn } from '@/components/a11y'
import { t } from '@/i18n'
import { Emote } from '@/features/emotes'
import { Resolution } from './Resolution'
import { Clipboard } from './Clipboard'
import { Player } from './Player'
import { useFullscreen } from './useFullscreen'
import { useClipboardSync, canReadClipboard } from './useClipboardSync'
import { usePip } from './usePip'
import './video.scss'

export function Stage({ hideControls, extraControls }: { hideControls: boolean; extraControls: boolean }) {
  const { state, app: a } = useNeko()
  const player = useRef<HTMLDivElement>(null)
  const [menu, setMenu] = useState<null | 'resolution' | 'clipboard'>(null)
  const { fullscreen, request: requestFullscreen } = useFullscreen(player)
  const syncClipboard = useClipboardSync()
  const pip = usePip()

  useEffect(() => (menu ? closeOn(() => setMenu(null)) : undefined), [menu])

  const admin = client.isAdmin
  const hosting = client.controlling
  const hosted = state.control.host_id !== null
  const implicit = state.settings.implicit_hosting
  const controlLocked = state.settings.locked_controls && !admin
  const { playing, playable, muted, mutedByAutoplay } = state.video
  const open = (m: 'resolution' | 'clipboard') => (e: React.MouseEvent) => (
    e.stopPropagation(),
    setMenu(menu === m ? null : m)
  )
  const extra = extraControls ? '' : 'extra-control'

  return (
    <div className="video">
      <div ref={player} className="player">
        <div className="player-container" onMouseEnter={syncClipboard}>
          <Player />
          <div className="emotes">
            {Object.entries(a.emotes).map(([id, type]) => (
              <Emote key={id} id={id} type={type} />
            ))}
          </div>
          {!playing && playable ? (
            <div
              className="player-overlay"
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
              <div className="player-overlay" {...a11y('Unmute')} onClick={() => client.unmute()}>
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
                <i {...a11y('Screen resolution')} onClick={open('resolution')} className="fas fa-desktop" />
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
                <i {...a11y('Clipboard')} onClick={open('clipboard')} className="fas fa-clipboard" />
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
                {...a11y('Keyboard')}
                onMouseDown={(e) => e.preventDefault()} // tapping the button must not take the focus the keyboard needs
                onClick={() => client.mobileKeyboardToggle()}
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
