import { Component, useEffect, useRef, useState, type FormEvent, type ReactNode } from 'react'
import {
  useNeko,
  actions,
  t,
  isLocked,
  setSetting,
  remember,
  client,
  tell,
  a11y,
  closeOn,
  type LockResource,
} from './neko'
import { Video } from './Video'
import { Members, MemberMenu, RoomMenu, Controls, Emotes } from './Room'
import { Side } from './Side'
import logo from '../assets/images/logo.svg'
import './styles/app.scss'
import './styles/header.scss'
import './styles/connect.scss'
import './styles/unsupported.scss'
import './styles/about.scss'
import { version } from '../../package.json'

const params = new URL(location.href).searchParams
// ?pwd= invite (e.g. neko-rooms links): used for the first login attempt only, like the legacy
// client; after that the form asks for a password, so a stale invite is not a dead end
let invite = params.get('pwd')
const loginOnce = (user: string, password: string) =>
  actions
    .login(user, password)
    .catch((err) => tell(t('connect.error'), err.message))
    .finally(() => (invite = null))
const cast = !!params.get('cast')
const videoOnly = cast || !!params.get('embed')

if (cast) setSetting('chat_sound', false)

export function App() {
  const { state, app } = useNeko()

  if (typeof RTCPeerConnection === 'undefined') {
    return (
      <div className="unsupported">
        <div className="window">
          <Logo />
          <div className="message">
            <span>{t('unsupported')}</span>
          </div>
        </div>
      </div>
    )
  }

  const connected = state.connection.status === 'connected'
  return (
    <div id="neko" className={!videoOnly && app.side ? 'expanded' : ''}>
      <main className="neko-main">
        {!videoOnly && (
          <div className="header-container">
            <Header />
          </div>
        )}
        <div className="video-container">
          <Video hideControls={cast} extraControls={videoOnly} />
        </div>
        {!videoOnly && (
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
        )}
      </main>
      {!videoOnly && <MemberMenu />}
      {!videoOnly && app.side && <Side />}
      {!connected && <Connect />}
      {!videoOnly && <Toasts />}
      {app.about && <About />}
      <Dialog />
    </div>
  )
}

function Logo() {
  return (
    <>
      <img src={logo} alt="n.eko" />
      <span>
        <b>n</b>.eko
      </span>
    </>
  )
}

function Header() {
  const { client, app } = useNeko()
  const admin = client.isAdmin
  const [read, setRead] = useState(app.texts)

  const lock = (r: LockResource, icon: string) => {
    const locked = isLocked(r)
    const tip = admin
      ? t(`locks.${r}.${locked ? 'unlock' : 'lock'}`)
      : t(`locks.${r}.${locked ? 'locked' : 'unlocked'}`)
    return (
      <li>
        <i
          className={`fas ${icon}${admin ? '' : ' disabled'}${locked ? ' locked' : ''}`}
          {...a11y(tip)}
          aria-disabled={!admin}
          onClick={() => admin && actions.toggleLock(r)}
        />
      </li>
    )
  }

  return (
    <div className="header">
      <a
        href="https://github.com/m1k1o/neko"
        title="Github repository"
        target="_blank"
        rel="noreferrer"
        className="neko"
      >
        <Logo />
      </a>
      <ul className="menu">
        {lock('control', 'fa-mouse')}
        {lock('login', isLocked('login') ? 'fa-lock' : 'fa-lock-open')}
        {app.files?.enabled && lock('file_transfer', 'fa-file')}
        <li>
          {!app.side && read !== app.texts && <span className="badge">&bull;</span>}
          <i
            className="fas fa-bars toggle"
            {...a11y('Toggle side panel')}
            aria-expanded={app.side}
            onClick={() => {
              app.side = !app.side
              remember('side', app.side)
              setRead(app.texts)
            }}
          />
        </li>
      </ul>
    </div>
  )
}

function Connect() {
  const { state } = useNeko()
  const [oauth, setOauth] = useState<{
    enabled?: boolean
    name?: string
    login_url?: string
    password_login_enabled?: boolean
  }>({})
  const [displayname, setDisplayname] = useState(params.get('usr') || '')
  const [password, setPassword] = useState('')

  useEffect(() => {
    client.api
      .req<typeof oauth>('GET', '/oauth/config')
      .then(setOauth)
      .catch(() => {})
    // ?usr=&pwd= (e.g. neko-rooms links) log in straight away, then leave the URL
    if (invite !== null || params.has('usr') || params.has('token')) {
      const url = new URL(location.href)
      for (const k of ['pwd', 'usr', 'token']) url.searchParams.delete(k)
      history.replaceState(null, '', url)
    }
    if (invite && params.get('usr') && !client.state.authenticated) {
      loginOnce(params.get('usr')!, invite)
    }
  }, [])

  const login = (e: FormEvent) => {
    e.preventDefault()
    if (!displayname) return tell(t('connect.error'), t('connect.empty_displayname'))
    loginOnce(displayname, invite ?? password)
  }

  const connecting = state.connection.status === 'connecting'
  const passwordLogin = oauth.password_login_enabled !== false
  return (
    <div className="connect">
      <div className="window">
        <div className="logo">
          <Logo />
        </div>
        {connecting ? (
          <div className="loader">
            <div className="bounce1"></div>
            <div className="bounce2"></div>
          </div>
        ) : state.authenticated ? (
          // logged in but disconnected (network drop, server restart)
          <form className="message" onSubmit={(e) => (e.preventDefault(), client.connect())}>
            <span>{t('connection.disconnected')}</span>
            <button type="submit">{t('connect.connect')}</button>
            <button type="button" onClick={actions.logout}>
              {t('logout')}
            </button>
          </form>
        ) : (
          <form className="message" onSubmit={login}>
            <span>{invite === null ? t('connect.login_title') : t('connect.invitation_title')}</span>
            {passwordLogin && (
              <input
                type="text"
                placeholder={t('connect.displayname')}
                value={displayname}
                onChange={(e) => setDisplayname(e.target.value)}
                autoFocus
              />
            )}
            {passwordLogin && invite === null && (
              <input
                type="password"
                placeholder={t('connect.password')}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            )}
            {passwordLogin && <button type="submit">{t('connect.connect')}</button>}
            {oauth.enabled && invite === null && (
              <button
                type="button"
                className="oauth-login"
                onClick={() => oauth.login_url && location.assign(oauth.login_url)}
              >
                {oauth.name || 'OAuth'} Login
              </button>
            )}
          </form>
        )}
      </div>
    </div>
  )
}

function Toasts() {
  const { app } = useNeko()
  return (
    <div className="toasts">
      {app.toasts.map((n) => (
        <div key={n.id} className={`toast ${n.kind}`} role="status">
          <div className="toast-title">{n.title}</div>
          {n.text && <div className="toast-text">{n.text}</div>}
        </div>
      ))}
    </div>
  )
}

const DIALOG_ICON = { warning: 'fa-exclamation-triangle', error: 'fa-times-circle', info: 'fa-info-circle' }

function Dialog() {
  const { app } = useNeko()
  const ref = useRef<HTMLDialogElement>(null)
  const d = app.dialog

  useEffect(() => {
    if (d && !ref.current!.open) ref.current!.showModal()
    if (!d && ref.current!.open) ref.current!.close()
  }, [d])

  const done = (ok: boolean) => {
    const resolve = app.dialog?.resolve
    app.dialog = null
    resolve?.(ok)
  }

  return (
    <dialog ref={ref} className="neko-dialog" onCancel={(e) => (e.preventDefault(), done(false))}>
      {d && (
        <>
          <i className={`icon ${d.icon} fas ${DIALOG_ICON[d.icon]}`} aria-hidden="true" />
          <h2>{d.title}</h2>
          {d.text && <p>{d.text}</p>}
          <div className="actions">
            <button className="confirm" autoFocus onClick={() => done(true)}>
              {d.cancel ? t('context.confirm.button_yes') : t('connection.button_confirm')}
            </button>
            {d.cancel && (
              <button className="cancel" onClick={() => done(false)}>
                {t('context.confirm.button_cancel')}
              </button>
            )}
          </div>
        </>
      )}
    </dialog>
  )
}

function About() {
  const { app } = useNeko()
  useEffect(() => closeOn(() => (app.about = false)), [app])
  return (
    <div className="about" onClick={(e) => e.target === e.currentTarget && (app.about = false)}>
      <div className="window" role="dialog" aria-label="About n.eko">
        <div className="about-content">
          <div className="logo">
            <Logo />
          </div>
          <p>A self hosted virtual browser that runs in docker and uses WebRTC.</p>
          <p className="links">
            <a href="https://github.com/m1k1o/neko" target="_blank" rel="noopener noreferrer">
              <i className="fab fa-github" /> m1k1o/neko
            </a>
            <a href="https://neko.m1k1o.net/" target="_blank" rel="noopener noreferrer">
              <i className="fas fa-book" /> Documentation
            </a>
          </p>
          <p className="version">client {version}</p>
          <button onClick={() => (app.about = false)}>{t('connection.button_confirm')}</button>
        </div>
      </div>
    </div>
  )
}

// a render bug should not leave a blank page
export class ErrorBoundary extends Component<{ children: ReactNode }, { error?: Error }> {
  state: { error?: Error } = {}

  static getDerivedStateFromError(error: Error) {
    return { error }
  }

  render() {
    if (!this.state.error) return this.props.children
    return (
      <div className="unsupported">
        <div className="window">
          <div className="logo">
            <Logo />
          </div>
          <div className="message">
            <span>Something went wrong: {this.state.error.message}</span>
            <button onClick={() => location.reload()}>Reload</button>
          </div>
        </div>
      </div>
    )
  }
}
