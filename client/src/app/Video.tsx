import { useEffect, useRef, useState } from 'react'
import { useNeko, client, app, actions, a11y, closeOn, t } from './neko'
import './styles/video.scss'
import './styles/emote.scss'
import './styles/resolution.scss'
import './styles/clipboard.scss'

// Firefox reports readText but hangs; Safari needs a gesture per read -> use the textarea fallback there
const ua = navigator.userAgent
const canReadClipboard =
  typeof navigator.clipboard?.readText === 'function' &&
  !ua.includes('Firefox') &&
  !(ua.includes('Safari') && !ua.includes('Chrome') && !ua.includes('Chromium'))
const canPip = typeof document.createElement('video').requestPictureInPicture === 'function'

async function syncClipboard() {
  if (!canReadClipboard || !client.controlling || !document.hasFocus()) return
  try {
    const text = await navigator.clipboard.readText()
    if (text !== client.state.control.clipboard?.text) client.send('clipboard/set', { text })
  } catch {}
}

export function Video({ hideControls, extraControls }: { hideControls: boolean; extraControls: boolean }) {
  const { state, app: a } = useNeko()
  const mount = useRef<HTMLDivElement>(null)
  const player = useRef<HTMLDivElement>(null)
  const [fullscreen, setFullscreen] = useState(false)
  const [menu, setMenu] = useState<null | 'resolution' | 'clipboard'>(null)

  useEffect(() => {
    client.mount(mount.current!)
    const onFs = () => {
      const fs = !!document.fullscreenElement
      setFullscreen(fs)
      // keyboard lock lets Esc/Alt+Tab etc. reach the remote while fullscreen
      const kb = (navigator as any).keyboard
      fs ? kb?.lock?.().catch(() => {}) : kb?.unlock?.()
    }
    document.addEventListener('fullscreenchange', onFs)
    window.addEventListener('focus', syncClipboard)
    return () => {
      client.unmount()
      document.removeEventListener('fullscreenchange', onFs)
      window.removeEventListener('focus', syncClipboard)
    }
  }, [])

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

  const requestFullscreen = () => {
    // iOS only allows fullscreen on the video element itself
    if (player.current?.requestFullscreen) player.current.requestFullscreen().catch(() => {})
    else (client.video as any)?.webkitEnterFullscreen?.()
  }

  return (
    <div className="video">
      <div ref={player} className="player">
        <div className="player-container" onMouseEnter={syncClipboard}>
          <div ref={mount} className="neko-mount" />
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
            {canPip && (
              <li>
                <i
                  {...a11y('Picture-in-Picture')}
                  onClick={() => client.video?.requestPictureInPicture().catch(() => {})}
                  className="fas fa-external-link-alt"
                />
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

function Resolution({ onPick }: { onPick: () => void }) {
  const { state } = useNeko()
  const { width, height, rate } = state.screen.size
  return (
    <ul className="resolution" role="menu" style={{ top: 50, right: 50 }} onClick={(e) => e.stopPropagation()}>
      {state.screen.configurations.map((c, i) => (
        <li
          key={i}
          className={c.width === width && c.height === height && c.rate === rate ? 'active' : ''}
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

function Clipboard() {
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

const rnd = (a: number, b: number) => a + Math.random() * (b - a)

// seven copies float up and fade, like the legacy anime.js version
function Emote({ id, type }: { id: string; type: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const anims = [...ref.current!.children].map((el, i) => {
      const side = (odd: boolean) => (odd ? rnd(0, 50) : rnd(-50, 0)) + '%'
      return (el as HTMLElement).animate(
        [
          { left: side(!!(i % 2)), top: '0%', opacity: 0, transform: 'rotate(0deg)' },
          { left: side(!!(i % 2)), opacity: 1, offset: 0.33 },
          { left: side(!(i % 2)), opacity: 0.5, offset: 0.66 },
          { left: side(!!(i % 2)), top: rnd(-600, -200) + '%', opacity: 0, transform: `rotate(${rnd(-35, 35)}deg)` },
        ],
        { duration: rnd(1000, 2000), easing: 'ease-in-out', fill: 'forwards' },
      ).finished
    })
    Promise.all(anims).then(() => delete app.state.emotes[id])
  }, [id])
  return (
    <div ref={ref} className="emote-container">
      {Array.from({ length: 7 }, (_, i) => (
        <div key={i} className={`emote ${type}`} />
      ))}
    </div>
  )
}
