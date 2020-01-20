package webrtc

import (
	"fmt"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/config"
	"n.eko.moe/neko/internal/event"
	"n.eko.moe/neko/internal/gst"
	"n.eko.moe/neko/internal/hid"
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
)

func New(sessions *session.SessionManager, conf *config.WebRTC) *WebRTCManager {
	logger := log.With().Str("module", "webrtc").Logger()
	engine := webrtc.MediaEngine{}
	engine.RegisterDefaultCodecs()

	setings := webrtc.SettingEngine{
		LoggerFactory: loggerFactory{
			logger: logger,
		},
	}

	return &WebRTCManager{
		logger:   logger,
		engine:   engine,
		setings:  setings,
		api:      webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(setings)),
		cleanup:  time.NewTicker(1 * time.Second),
		shutdown: make(chan bool),
		sessions: sessions,
		conf:     conf,
		config: webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		},
	}
}

type WebRTCManager struct {
	logger        zerolog.Logger
	engine        webrtc.MediaEngine
	setings       webrtc.SettingEngine
	config        webrtc.Configuration
	sessions      *session.SessionManager
	api           *webrtc.API
	video         *webrtc.Track
	audio         *webrtc.Track
	videoPipeline *gst.Pipeline
	audioPipeline *gst.Pipeline
	cleanup       *time.Ticker
	conf          *config.WebRTC
	shutdown      chan bool
}

func (m *WebRTCManager) Start() {

	hid.Display(m.conf.Display)

	switch m.conf.VideoCodec {
	case "vp8":
		if err := m.createVideoTrack(webrtc.DefaultPayloadTypeVP8); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	case "vp9":
		if err := m.createVideoTrack(webrtc.DefaultPayloadTypeVP9); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	case "h264":
		if err := m.createVideoTrack(webrtc.DefaultPayloadTypeH264); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	default:
		m.logger.Panic().Err(errors.Errorf("unknown video codec %s", m.conf.AudioCodec)).Msg("unable to start webrtc manager")
	}

	switch m.conf.AudioCodec {
	case "opus":
		if err := m.createAudioTrack(webrtc.DefaultPayloadTypeOpus); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	case "g722":
		if err := m.createAudioTrack(webrtc.DefaultPayloadTypeG722); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	case "pcmu":
		if err := m.createAudioTrack(webrtc.DefaultPayloadTypePCMU); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	case "pcma":
		if err := m.createAudioTrack(webrtc.DefaultPayloadTypePCMA); err != nil {
			m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
		}
	default:
		m.logger.Panic().Err(errors.Errorf("unknown audio codec %s", m.conf.AudioCodec)).Msg("unable to start webrtc manager")
	}

	m.videoPipeline.Start()
	m.audioPipeline.Start()

	go func() {
		defer func() {
			m.logger.Info().Msg("shutdown")
		}()

		for {
			select {
			case <-m.shutdown:
				return
			case <-m.cleanup.C:
				hid.Check(time.Second * 10)
			}
		}
	}()

	m.sessions.OnHostCleared(func(id string) {
		hid.Reset()
	})

	m.sessions.OnCreated(func(id string, session *session.Session) {
		m.logger.Debug().Str("id", id).Msg("session created")
	})

	m.sessions.OnDestroy(func(id string) {
		m.logger.Debug().Str("id", id).Msg("session destroyed")
	})

	// TODO: log resolution, bit rate and codec parameters
	m.logger.Info().
		Str("video_display", m.conf.Display).
		Str("video_codec", m.conf.VideoCodec).
		Str("audio_device", m.conf.Device).
		Str("audio_codec", m.conf.AudioCodec).
		Msgf("webrtc streaming")
}

func (m *WebRTCManager) Shutdown() error {
	m.logger.Info().Msgf("webrtc shutting down")

	m.cleanup.Stop()
	m.shutdown <- true
	m.videoPipeline.Stop()
	m.audioPipeline.Stop()
	return nil
}

func (m *WebRTCManager) CreatePeer(id string, sdp string) error {
	session, ok := m.sessions.Get(id)
	if !ok {
		return fmt.Errorf("invalid session id %s", id)
	}

	peer, err := m.api.NewPeerConnection(m.config)
	if err != nil {
		return err
	}

	if _, err := peer.AddTransceiverFromTrack(m.video, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return err
	}

	if _, err := peer.AddTransceiverFromTrack(m.audio, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return err
	}

	peer.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	})

	answer, err := peer.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err = peer.SetLocalDescription(answer); err != nil {
		return err
	}

	if err := session.Send(message.Signal{
		Event: event.SIGNAL_ANSWER,
		SDP:   answer.SDP,
	}); err != nil {
		return err
	}

	peer.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if err = m.handle(id, msg); err != nil {
				m.logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	peer.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
		switch connectionState {
		case webrtc.PeerConnectionStateDisconnected:
		case webrtc.PeerConnectionStateFailed:
			m.logger.Info().Str("id", id).Msg("peer disconnected")
			m.sessions.Destroy(id)
			break
		case webrtc.PeerConnectionStateConnected:
			m.logger.Info().Str("id", id).Msg("peer connected")
			break
		}
	})

	m.sessions.SetPeer(id, peer)

	return nil
}
