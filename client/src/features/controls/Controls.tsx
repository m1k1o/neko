import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin, selectSession } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { actions } from '@/state/actions'
import { client } from '@/state/client'
import { tell } from '@/state/dialogs'
import { a11y } from '@/components/a11y'
import './controls.scss'

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
    <ul className="neko-controls flex items-center justify-center">
      {!implicit && (!controlLocked || hosting) && (
        <li>
          <i
            className={[
              !hosted || hosting ? '' : 'disabled',
              !hosted && !hosting ? 'faded' : '',
              shake && !hosting ? 'shake' : '',
              'fas fa-keyboard request',
            ].join(' ')}
            data-testid="control-request"
            {...a11y(hosting ? t('controls.release') : t('controls.request'))}
            onClick={() => playable && actions.toggleControl()}
          />
        </li>
      )}
      {implicit && (
        <li className="no-pointer">
          <i
            className={`${controlLocked ? 'disabled ' : ''}fas fa-mouse-pointer`}
            data-testid="control-implicit"
            title={t(controlLocked ? 'controls.hasnot' : 'controls.has')}
          />
        </li>
      )}
      {(implicit || !controlLocked || hosting) && (
        <li>
          <label className="switch" title={hosting ? t(locked ? 'controls.unlock' : 'controls.lock') : ''}>
            <input
              type="checkbox"
              aria-label={t(locked ? 'controls.unlock' : 'controls.lock')}
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
            data-testid="mic"
            data-on={micOn || undefined}
            {...a11y(t(micOn ? 'controls.mic_off' : 'controls.mic_on'))}
            onClick={toggleMic}
          />
        </li>
      )}
      <li>
        <div className="volume">
          <i
            className={`fas ${volume === 0 || muted ? 'fa-volume-mute' : 'fa-volume-up'}`}
            data-testid="mute"
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
