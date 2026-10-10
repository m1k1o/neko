import { useEffect, useReducer, useState, type FormEvent } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { actions } from '@/state/actions'
import { client } from '@/state/client'
import { tell } from '@/state/dialogs'
import { i18n } from '@/i18n'
import logo from '@/assets/images/logo.svg'

const input =
  'my-[5px] rounded-[5px] border-0 bg-background-tertiary px-2 py-1.5 leading-5 text-text-normal selection:bg-text-link placeholder:text-[#757575]'
const button =
  'my-[5px] cursor-pointer rounded-[5px] border-0 bg-style-primary p-1 text-center leading-7.5 font-bold text-text-normal uppercase'

const params = new URL(location.href).searchParams
// ?pwd= invite (e.g. neko-rooms links): used for the first login attempt only, like the legacy
// client; after that the form asks for a password, so a stale invite is not a dead end
let invite = params.get('pwd')
const loginOnce = (user: string, password: string) =>
  actions
    .login(user, password)
    .catch((err) => tell(i18n.t('connect.error'), err.message))
    .finally(() => (invite = null))

// the login screen, and the "Connect" screen of a logged-in viewer without a connection
export function Connect() {
  const { t } = useTranslation()
  const { status, authenticated } = useStore(
    client.store,
    useShallow((s) => ({ status: s.connection.status, authenticated: s.authenticated })),
  )
  const [oauth, setOauth] = useState<{
    enabled?: boolean
    name?: string
    login_url?: string
    password_login_enabled?: boolean
  }>({})
  const [displayname, setDisplayname] = useState(params.get('usr') || '')
  const [password, setPassword] = useState('')
  // a login attempt that is over has used up the invite: show the form with the password field
  const [, settled] = useReducer((n: number) => n + 1, 0)

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
    // always, even if a saved session is being resumed at the same time: the link wins
    // (login() drops the resumed websocket), and this must not depend on which answer is first
    if (invite && params.get('usr')) loginOnce(params.get('usr')!, invite).finally(settled)
  }, [])

  const login = (e: FormEvent) => {
    e.preventDefault()
    if (!displayname) return tell(t('connect.error'), t('connect.empty_displayname'))
    loginOnce(displayname, invite ?? password).finally(settled)
  }

  const connecting = status === 'connecting'
  const passwordLogin = oauth.password_login_enabled !== false
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-background-floating/80" data-testid="connect">
      <div className="w-[320px] rounded-[5px] bg-background-secondary p-2.5">
        <div className="flex w-full cursor-pointer items-center justify-center">
          <img src={logo} alt="n.eko" className="mr-2.5 h-[90px]" />
          <span className="text-[30px] leading-14">
            <b className="font-black">n</b>.eko
          </span>
        </div>
        {connecting ? (
          <div className="relative mx-auto h-[90px] w-[90px]">
            <div className="absolute top-0 left-0 h-full w-full animate-loader rounded-full bg-style-primary opacity-60"></div>
            <div className="absolute top-0 left-0 h-full w-full animate-loader rounded-full bg-style-primary opacity-60 [animation-delay:-1s]"></div>
          </div>
        ) : authenticated ? (
          // logged in but disconnected (network drop, server restart)
          <form className="flex flex-col" onSubmit={(e) => (e.preventDefault(), client.connect())}>
            <span className="block text-center leading-7.5 uppercase">{t('connection.disconnected')}</span>
            <button className={button} type="submit">
              {t('connect.connect')}
            </button>
            <button className={button} type="button" onClick={actions.logout}>
              {t('logout')}
            </button>
          </form>
        ) : (
          <form className="flex flex-col" onSubmit={login}>
            <span className="block text-center leading-7.5 uppercase">
              {invite === null ? t('connect.login_title') : t('connect.invitation_title')}
            </span>
            {passwordLogin && (
              <input
                className={input}
                type="text"
                placeholder={t('connect.displayname')}
                value={displayname}
                onChange={(e) => setDisplayname(e.target.value)}
                autoFocus
              />
            )}
            {passwordLogin && invite === null && (
              <input
                className={input}
                type="password"
                placeholder={t('connect.password')}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            )}
            {passwordLogin && (
              <button className={button} type="submit">
                {t('connect.connect')}
              </button>
            )}
            {oauth.enabled && invite === null && (
              <button
                className={cn(button, 'border border-style-primary bg-background-tertiary')}
                type="button"
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
