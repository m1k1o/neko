import { useEffect, useRef, useState } from 'react'
import { useNeko, actions, t, isMuted, client, app, name, ask, tell, langs, setLang, a11y } from './neko'
import './styles/members.scss'
import './styles/menu.scss'
import './styles/controls.scss'
import './styles/emotes.scss'
import './styles/context.scss'
import './styles/avatar.scss'

const EMOTES = [
  'anger',
  'bomb',
  'sleep',
  'explode',
  'sweat',
  'poo',
  'hundred',
  'alert',
  'punch',
  'wave',
  'okay',
  'thumbs-up',
  'clap',
  'prey',
  'celebrate',
  'flame',
  'goof',
  'love',
  'cool',
  'smerk',
  'worry',
  'ouch',
  'cry',
  'surprised',
  'quiet',
  'rage',
  'annoy',
  'steamed',
  'scared',
  'terrified',
  'sleepy',
  'dead',
  'happy',
  'roll-eyes',
  'thinking',
  'clown',
  'sick',
  'rofl',
  'drule',
  'sniff',
  'sus',
  'party',
  'odd',
  'hot',
  'cold',
  'blush',
  'sad',
]

export function Avatar({ seed, avatar, size }: { seed: string; avatar?: string; size: number }) {
  const [failed, setFailed] = useState(false)
  useEffect(() => setFailed(false), [avatar])

  let url = ''
  try {
    const u = new URL(avatar || '')
    if (u.protocol === 'http:' || u.protocol === 'https:') url = u.toString()
  } catch {}

  // deterministic pastel color from the name, same formula as the legacy client
  let a = 0,
    b = 0,
    c = 0
  for (let i = 0; i < seed.length; i++) {
    a += seed.charCodeAt(i) * 3
    b += seed.charCodeAt(i) * 5
    c += seed.charCodeAt(i) * 7
  }
  const bg = `rgb(${128 + (a % 128)},${128 + (b % 128)},${128 + (c % 128)})`

  return (
    <div
      className="avatar"
      style={{ width: size, height: size, lineHeight: size + 'px', fontSize: size / 2, backgroundColor: bg }}
    >
      {url && !failed ? (
        <img src={url} alt={seed} onError={() => setFailed(true)} />
      ) : (
        seed.substring(0, 2).toUpperCase()
      )}
    </div>
  )
}

// keyboard equivalent of right-click: the Menu key or Shift+F10 on a focused member
const menuKey = (e: React.KeyboardEvent, id: string) => {
  if (e.key !== 'ContextMenu' && !(e.shiftKey && e.key === 'F10')) return
  e.preventDefault()
  e.stopPropagation()
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  if (id !== client.state.session_id) app.state.menu = { x: r.left, y: r.bottom, id }
}

export const openMenu = (e: React.MouseEvent, id: string) => {
  e.preventDefault()
  e.stopPropagation()
  if (id !== client.state.session_id) app.state.menu = { x: e.clientX, y: e.clientY, id }
}

export function Members() {
  const { state } = useNeko()
  const me = client.session
  const host = state.control.host_id
  return (
    <div className="members">
      <div className="members-container">
        <ul className="members-list">
          {me && (
            <li>
              <div className={`member self${me.id === host ? ' host' : ''}`}>
                <Avatar seed={me.profile.name} avatar={me.profile.avatar} size={50} />
              </div>
            </li>
          )}
          {Object.values(state.sessions)
            .filter((m) => m.id !== state.session_id && m.state.is_connected)
            .map((m) => (
              <li key={m.id} title={m.profile.name}>
                <div
                  className={`member${m.id === host ? ' host' : ''}${m.profile.is_admin ? ' admin' : ''}`}
                  tabIndex={0}
                  aria-label={m.profile.name}
                  aria-haspopup="menu"
                  onContextMenu={(e) => openMenu(e, m.id)}
                  onKeyDown={(e) => menuKey(e, m.id)}
                >
                  <Avatar seed={m.profile.name} avatar={m.profile.avatar} size={50} />
                </div>
              </li>
            ))}
        </ul>
      </div>
      <MemberMenu />
    </div>
  )
}

const confirmThen = (title: string, text: string, fn: () => void) => ask(title, text).then((ok) => ok && fn())

function MemberMenu() {
  const { app, state } = useNeko()
  const [bannable, setBannable] = useState(false)
  useEffect(() => {
    const close = () => (app.menu = null)
    document.addEventListener('click', close)
    return () => document.removeEventListener('click', close)
  }, [])

  const id = app.menu?.id
  const m = id ? state.sessions[id] : undefined
  // ban only sticks where the auth provider stores accounts, see actions.ban
  useEffect(() => {
    setBannable(false)
    if (id && client.isAdmin) actions.canBan(id).then(setBannable)
  }, [id])
  if (!app.menu || !m) return null

  const n = m.profile.name
  const admin = client.isAdmin
  const implicit = state.settings.implicit_hosting
  const isHost = id === state.control.host_id
  const muted = isMuted(id)
  const x = Math.min(app.menu.x, innerWidth - 170)
  const y = Math.min(app.menu.y, innerHeight - 250)

  return (
    <ul className="context" role="menu" aria-label={n} style={{ left: x, top: y }}>
      <li className="header">
        <div className="user">
          <Avatar seed={n} avatar={m.profile.avatar} size={25} />
          <strong>{n}</strong>
        </div>
      </li>
      <li className="seperator" />
      <li>
        <span
          {...a11y(t(app.ignored[id!] ? 'context.unignore' : 'context.ignore'), 'menuitem')}
          onClick={() => (app.ignored[id!] = !app.ignored[id!])}
        >
          {t(app.ignored[id!] ? 'context.unignore' : 'context.ignore')}
        </span>
      </li>
      {admin ? (
        <>
          <li>
            <span
              {...a11y(t(muted ? 'context.unmute' : 'context.mute'), 'menuitem')}
              onClick={() =>
                confirmThen(
                  t(`context.confirm.${muted ? 'unmute' : 'mute'}_title`, { name: n }),
                  t(`context.confirm.${muted ? 'unmute' : 'mute'}_text`, { name: n }),
                  () => actions.mute(id!, !muted),
                )
              }
            >
              {t(muted ? 'context.unmute' : 'context.mute')}
            </span>
          </li>
          {!implicit && isHost && (
            <>
              <li>
                <span {...a11y(t('context.release'), 'menuitem')} onClick={actions.reset}>
                  {t('context.release')}
                </span>
              </li>
              <li>
                <span {...a11y(t('context.take'), 'menuitem')} onClick={actions.take}>
                  {t('context.take')}
                </span>
              </li>
            </>
          )}
          {!implicit && !isHost && (
            <li>
              <span {...a11y(t('context.give'), 'menuitem')} onClick={() => actions.give(id!)}>
                {t('context.give')}
              </span>
            </li>
          )}
        </>
      ) : (
        client.controlling &&
        !implicit && (
          <li>
            <span {...a11y(t('context.give'), 'menuitem')} onClick={() => actions.give(id!)}>
              {t('context.give')}
            </span>
          </li>
        )
      )}
      {admin && !m.profile.is_admin && (
        <>
          <li className="seperator" />
          <li>
            <span
              style={{ color: '#f04747' }}
              {...a11y(t('context.kick'), 'menuitem')}
              onClick={() =>
                confirmThen(
                  t('context.confirm.kick_title', { name: n }),
                  t('context.confirm.kick_text', { name: n }),
                  () => actions.kick(id!),
                )
              }
            >
              {t('context.kick')}
            </span>
          </li>
          {bannable && (
            <li>
              <span
                style={{ color: '#f04747' }}
                {...a11y(t('context.ban'), 'menuitem')}
                onClick={() =>
                  confirmThen(
                    t('context.confirm.ban_title', { name: n }),
                    t('context.confirm.ban_text', { name: n }),
                    () => actions.ban(id!),
                  )
                }
              >
                {t('context.ban')}
              </span>
            </li>
          )}
        </>
      )}
    </ul>
  )
}

export function RoomMenu() {
  useNeko()
  return (
    <ul className="room-settings">
      <li>
        <i className="fas fa-question-circle" {...a11y('About n.eko')} onClick={() => (app.state.about = true)} />
      </li>
      <li>{client.isAdmin && <i className="fas fa-shield-alt" title={t('admin_loggedin')} />}</li>
      <li>
        <select
          value={app.state.lang}
          onChange={(e) => setLang(e.target.value as typeof app.state.lang)}
          aria-label="Language"
        >
          {langs.map((l) => (
            <option key={l} value={l}>
              {l}
            </option>
          ))}
        </select>
      </li>
    </ul>
  )
}

export function Controls() {
  const { state } = useNeko()
  const [shake, setShake] = useState(false)
  const mic = useRef<{ track: MediaStreamTrack; sender: RTCRtpSender } | null>(null)
  const [micOn, setMicOn] = useState(false)

  const admin = client.isAdmin
  const hosting = client.controlling
  const hosted = state.control.host_id !== null
  const implicit = state.settings.implicit_hosting
  const controlLocked = state.settings.locked_controls && !admin
  const locked = state.control.locked && hosting
  const { playable, playing, muted, volume } = state.video
  const micAllowed = hosting && !!client.session?.profile.can_share_media

  // clicking the video without control shakes the keyboard icon as a hint
  useEffect(() => {
    const hint = () => {
      if (client.controlling || client.state.settings.implicit_hosting) return
      setShake(true)
      setTimeout(() => setShake(false), 5000)
    }
    return client.events.on('overlay.click', hint)
  }, [])

  const micOff = () => {
    if (!mic.current) return
    client.removeTrack(mic.current.sender)
    mic.current.track.stop()
    mic.current = null
    setMicOn(false)
  }
  // drop the mic when control is lost, so the next host gets the audio input
  useEffect(() => {
    if (!micAllowed) micOff()
  }, [micAllowed])

  const toggleMic = async () => {
    if (mic.current) return micOff()
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const track = stream.getAudioTracks()[0]
      mic.current = { track, sender: client.addTrack(track, stream) }
      setMicOn(true)
    } catch (err: any) {
      tell(t('controls.mic_error'), err.message)
    }
  }

  return (
    <ul className="neko-controls">
      {!implicit && (!controlLocked || hosting) && (
        <li>
          <i
            className={[
              !hosted || hosting ? '' : 'disabled',
              !hosted && !hosting ? 'faded' : '',
              shake && !hosting ? 'shake' : '',
              'fas fa-keyboard request',
            ].join(' ')}
            {...a11y(hosting ? t('controls.release') : t('controls.request'))}
            onClick={() => playable && actions.toggleControl()}
          />
        </li>
      )}
      {implicit && (
        <li className="no-pointer">
          <i
            className={`${controlLocked ? 'disabled ' : ''}fas fa-mouse-pointer`}
            title={t(controlLocked ? 'controls.hasnot' : 'controls.has')}
          />
        </li>
      )}
      {(implicit || !controlLocked || hosting) && (
        <li>
          <label className="switch" title={hosting ? t(locked ? 'controls.unlock' : 'controls.lock') : ''}>
            <input
              type="checkbox"
              checked={locked}
              disabled={!hosting || (implicit && controlLocked)}
              onChange={(e) => (e.target.checked ? client.lock() : client.unlock())}
            />
            <span />
          </label>
        </li>
      )}
      <li>
        <i
          className={`${playable ? '' : 'disabled '}fas ${playing ? 'fa-pause-circle' : 'fa-play-circle'} play`}
          {...a11y(playing ? 'Pause' : 'Play')}
          onClick={() => playable && (playing ? client.pause() : client.play().catch(() => {}))}
        />
      </li>
      {micAllowed && (
        <li>
          <i
            className={`fas ${micOn ? 'fa-microphone' : 'fa-microphone-slash faded'}`}
            {...a11y(t(micOn ? 'controls.mic_off' : 'controls.mic_on'))}
            onClick={toggleMic}
          />
        </li>
      )}
      <li>
        <div className="volume">
          <i
            className={`fas ${volume === 0 || muted ? 'fa-volume-mute' : 'fa-volume-up'}`}
            {...a11y(muted ? 'Unmute' : 'Mute')}
            onClick={() => (muted ? client.unmute() : client.mute())}
          />
          <input
            type="range"
            min="0"
            max="100"
            aria-label="Volume"
            value={Math.round(volume * 100)}
            onChange={(e) => client.setVolume(Number(e.target.value) / 100)}
          />
        </div>
      </li>
    </ul>
  )
}

export function Emotes() {
  useNeko()
  const [recent, setRecent] = useState<string[]>(() => {
    try {
      return JSON.parse(localStorage.getItem('emote_recent') || '[]')
    } catch {
      return []
    }
  })
  const [picker, setPicker] = useState<{ x: number; y: number } | null>(null)
  const repeat = useRef(0)

  useEffect(() => {
    if (!picker) return
    const close = () => setPicker(null)
    document.addEventListener('click', close)
    return () => document.removeEventListener('click', close)
  }, [picker])

  const pick = (emote: string) => {
    if (!recent.includes(emote)) {
      const next = [...recent.slice(-4), emote]
      setRecent(next)
      try {
        localStorage.setItem('emote_recent', JSON.stringify(next))
      } catch {}
    }
    actions.sendEmote(emote)
  }
  const stop = () => clearInterval(repeat.current)
  // press and hold keeps sending, like the legacy client
  const start = (emote: string) => {
    actions.sendEmote(emote)
    stop()
    repeat.current = window.setInterval(() => actions.sendEmote(emote), 350)
  }

  if (isMuted()) return null
  return (
    <div className="emotes-bar" onMouseLeave={stop} onMouseUp={stop}>
      <ul>
        {recent.map((e) => (
          <li key={e}>
            <div
              className={`emote ${e}`}
              {...a11y(e)}
              onClick={(ev) => ev.detail === 0 && actions.sendEmote(e)}
              onMouseDown={(ev) => (ev.preventDefault(), start(e))}
            />
          </li>
        ))}
        <li>
          <i
            className="fas fa-grin-beam"
            {...a11y('Emotes')}
            onClick={(e) => (e.stopPropagation(), setPicker({ x: e.clientX, y: e.clientY }))}
          />
        </li>
      </ul>
      {picker && (
        <ul
          className="context"
          style={{ left: Math.min(picker.x, innerWidth - 260), top: Math.max(picker.y - 260, 10) }}
        >
          {EMOTES.filter((e) => !recent.includes(e)).map((e) => (
            <li key={e}>
              <div className={`emote ${e}`} {...a11y(e, 'menuitem')} onClick={() => pick(e)} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export { name }
