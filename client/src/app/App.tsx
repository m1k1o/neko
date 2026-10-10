import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { setSetting } from '@/state/settings'
import { useSlot } from '@/plugins'
import { WebRTCTransport } from '@m1k1o/neko'
import { Logo } from '@/components/Logo'
import { Dialog } from '@/components/Dialog'
import { Toasts } from '@/components/Toasts'
import { Header } from '@/layout/Header'
import { Side } from '@/layout/Side'
import { RoomBar } from '@/layout/RoomBar'
import { Stage } from '@/features/video'
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
  const { t } = useTranslation()
  const connected = useStore(client.store, (s) => s.connection.status === 'connected')
  const { side, about } = useStore(
    app,
    useShallow((s) => ({ side: s.side, about: s.about })),
  )
  const memberMenu = useSlot('member.menu')

  if (!WebRTCTransport.supported()) {
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

  return (
    <div id="neko" className={!videoOnly && side ? 'expanded' : ''}>
      <main className="neko-main">
        {!videoOnly && (
          <div className="header-container" data-testid="header">
            <Header />
          </div>
        )}
        <div className="video-container">
          <Stage hideControls={cast} extraControls={videoOnly} />
        </div>
        {!videoOnly && <RoomBar />}
      </main>
      {!videoOnly && <MemberMenu items={memberMenu} />}
      {!videoOnly && side && <Side />}
      {!connected && <Connect />}
      {!videoOnly && <Toasts />}
      {about && <About />}
      <Dialog />
    </div>
  )
}
