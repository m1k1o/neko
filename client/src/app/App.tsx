import './boot'
import { useNeko } from '@/state/hooks'
import { setSetting } from '@/state/settings'
import { t } from '@/i18n'
import { memberMenu } from '@/plugins'
import { Logo } from '@/components/Logo'
import { Dialog } from '@/components/Dialog'
import { Toasts } from '@/components/Toasts'
import { Header } from '@/layout/Header'
import { Side } from '@/layout/Side'
import { RoomBar } from '@/layout/RoomBar'
import { Video } from '@/features/video'
import { MemberMenu } from '@/features/members'
import { Connect } from '@/features/connect'
import { About } from '@/features/about'
import './app.scss'
import './unsupported.scss'

const params = new URL(location.href).searchParams
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
        {!videoOnly && <RoomBar />}
      </main>
      {!videoOnly && <MemberMenu items={memberMenu()} />}
      {!videoOnly && app.side && <Side />}
      {!connected && <Connect />}
      {!videoOnly && <Toasts />}
      {app.about && <About />}
      <Dialog />
    </div>
  )
}
