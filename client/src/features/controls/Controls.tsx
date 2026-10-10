import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin, selectSession } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { actions } from '@/state/actions'
import { client } from '@/state/client'
import { tell } from '@/state/dialogs'
import { a11y } from '@/components/a11y'

export function Controls() {
  const { t } = useTranslation()
  const { admin, hosting, hosted, implicit, lockedControls, localLock, video, connected, canShareMedia } = useStore(
    client.store,
    useShallow((s) => ({
      admin: selectIsAdmin(s),
      hosting: selectControlling(s),
      hosted: s.control.host_id !== null,
      implicit: s.settings.implicit_hosting,
      lockedControls: s.settings.locked_controls,
      localLock: s.control.locked,
      video: s.video,
      connected: s.connection.status === 'connected',
      canShareMedia: !!selectSession(s)?.profile.can_share_media,
    })),
  )
  const [shake, setShake] = useState(false)
  const mic = useRef<{ track: MediaStreamTrack; stop: () => void } | null>(null)
  const [micOn, setMicOn] = useState(false)

  const controlLocked = lockedControls && !admin
  const locked = localLock && hosting
  const { playable, playing, muted, volume } = video
  const micAllowed = hosting && canShareMedia
  const lockDisabled = !hosting || (implicit && controlLocked)

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
    mic.current.stop()
    mic.current.track.stop()
    mic.current = null
    setMicOn(false)
  }
  // drop the mic when control is lost (the next host gets the audio input) and whenever the
  // connection is re-established: the track was attached to the old peer connection
  useEffect(() => {
    if (!micAllowed || !connected) micOff()
  }, [micAllowed, connected])

  const toggleMic = async () => {
    if (mic.current) return micOff()
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const track = stream.getAudioTracks()[0]
      mic.current = { track, stop: client.shareMedia(stream) }
      setMicOn(true)
    } catch (err: any) {
      tell(t('controls.mic_error'), err.message)
    }
  }

  return (
    <ul className="flex items-center justify-center self-center text-[24px]">
      {!implicit && (!controlLocked || hosting) && (
        <li className="cursor-pointer">
          <i
            className={cn(
              'fas fa-keyboard px-[5px]',
              hosted && !hosting && 'text-style-error/40',
              !hosted && !hosting && 'text-text-normal/40',
              shake && !hosting && 'animate-shake',
            )}
            data-testid="control-request"
            {...a11y(hosting ? t('controls.release') : t('controls.request'))}
            onClick={() => playable && actions.toggleControl()}
          />
        </li>
      )}
      {implicit && (
        <li>
          <i
            className={cn('fas fa-mouse-pointer px-[5px]', controlLocked && 'text-style-error/40')}
            data-testid="control-implicit"
            title={t(controlLocked ? 'controls.hasnot' : 'controls.has')}
          />
        </li>
      )}
      {(implicit || !controlLocked || hosting) && (
        <li className="cursor-pointer">
          {/* the lock switch: the input, its track and its knob (a lock icon when it can be used) */}
          <label
            className="relative mx-[5px] block h-6 w-[42px]"
            title={hosting ? t(locked ? 'controls.unlock' : 'controls.lock') : ''}
          >
            <input
              type="checkbox"
              className="peer h-0 w-0 opacity-0"
              aria-label={t(locked ? 'controls.unlock' : 'controls.lock')}
              checked={locked}
              disabled={lockDisabled}
              onChange={(e) => (e.target.checked ? client.lock() : client.unlock())}
            />
            <span className="absolute inset-0 cursor-pointer rounded-[34px] bg-background-secondary transition-all duration-200 peer-checked:bg-style-primary" />
            <i
              className={cn(
                'absolute bottom-[3px] left-[3px] h-[18px] w-[18px] cursor-pointer rounded-full text-center text-[8px] leading-[18px] text-background-tertiary shadow-[0_2px_4px_rgba(0,0,0,0.3)] transition-all duration-300 peer-checked:translate-x-[18px]',
                lockDisabled ? 'bg-text-normal/40' : `fas bg-white ${locked ? 'fa-lock' : 'fa-lock-open'}`,
              )}
            />
          </label>
        </li>
      )}
      <li className="cursor-pointer">
        <i
          className={cn(
            `fas ${playing ? 'fa-pause-circle' : 'fa-play-circle'} px-[5px]`,
            !playable && 'text-style-error/40',
          )}
          {...a11y(playing ? 'Pause' : 'Play')}
          onClick={() => playable && (playing ? client.pause() : client.play().catch(() => {}))}
        />
      </li>
      {micAllowed && (
        <li className="cursor-pointer">
          <i
            className={cn('fas px-[5px]', micOn ? 'fa-microphone' : 'fa-microphone-slash text-text-normal/40')}
            data-testid="mic"
            data-on={micOn || undefined}
            {...a11y(t(micOn ? 'controls.mic_off' : 'controls.mic_on'))}
            onClick={toggleMic}
          />
        </li>
      )}
      <li className="cursor-pointer">
        <div className="flex items-center justify-center whitespace-nowrap">
          <i
            className={`fas ${volume === 0 || muted ? 'fa-volume-mute' : 'fa-volume-up'} px-[5px]`}
            data-testid="mute"
            {...a11y(muted ? 'Unmute' : 'Mute')}
            onClick={() => (muted ? client.unmute() : client.mute())}
          />
          <input
            type="range"
            className="slider h-5 w-[150px]"
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
