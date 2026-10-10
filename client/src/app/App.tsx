import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'
import { client } from '@/state/client'
import { setSetting } from '@/state/settings'
import { useSlot } from '@/plugins'
import { WebRTCTransport } from '@m1k1o/neko'
import { cn } from '@/lib/utils'
import logo from '@/assets/images/logo.svg'
import { Dialog } from '@/components/Dialog'
import { Toaster } from '@/components/ui/sonner'
import { Header } from '@/layout/Header'
import { Side } from '@/layout/Side'
import { RoomBar } from '@/layout/RoomBar'
import { Stage } from '@/features/video/Stage'
import { MemberMenu } from '@/features/members/MemberMenu'
import { Connect } from '@/features/connect/Connect'
import { About } from '@/features/about/About'

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
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="w-[320px] rounded-[5px] bg-background-secondary p-2.5">
          <div className="flex w-full items-center justify-center">
            <img src={logo} alt="n.eko" className="mr-2.5 h-[90px]" />
            <span className="text-[30px] leading-14">
              <b className="font-black">n</b>.eko
            </span>
          </div>
          <div className="flex flex-col">
            <span className="block text-center leading-7.5 uppercase">{t('unsupported')}</span>
          </div>
        </div>
      </div>
    )
  }

  const expanded = !videoOnly && side
  return (
    <div
      id="neko"
      className="tablet:relative tablet:max-h-none tablet:flex-col absolute inset-0 flex max-h-[100vh] max-w-[100vw]"
    >
      <main
        className={cn(
          'flex min-w-[360px] max-w-full grow flex-col overflow-auto',
          expanded ? 'tablet:portrait:h-[40vh] tablet:landscape:h-screen' : 'tablet:h-screen',
        )}
      >
        {!videoOnly && (
          <div className="flex h-menu shrink-0 bg-background-tertiary" data-testid="header">
            <Header />
          </div>
        )}
        {/* .video-container: index.html sizes it before the app renders (iOS) */}
        <div className="video-container flex max-w-full grow bg-black/40">
          <Stage hideControls={cast} extraControls={videoOnly} />
        </div>
        {!videoOnly && <RoomBar />}
      </main>
      {!videoOnly && <MemberMenu items={memberMenu} />}
      {expanded && <Side />}
      {!connected && <Connect />}
      {!videoOnly && <Toaster />}
      {about && <About />}
      <Dialog />
    </div>
  )
}
