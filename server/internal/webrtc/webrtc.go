package webrtc

import (
	"fmt"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/gst"
	"n.eko.moe/neko/internal/hid"
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/config"
)

func New(sessions types.SessionManager, config *config.WebRTC) *WebRTCManager {
	logger := log.With().Str("module", "webrtc").Logger()
	settings := webrtc.SettingEngine{
		LoggerFactory: loggerFactory{
			logger: logger,
		},
	}

	settings.SetLite(true)
	settings.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeUDP4})
	settings.SetEphemeralUDPPortRange(config.EphemeralMin, config.EphemeralMax)
	settings.SetNAT1To1IPs(config.NAT1To1IPs, webrtc.ICECandidateTypeHost)

	return &WebRTCManager{
		logger:   logger,
		settings: settings,
		cleanup:  time.NewTicker(1 * time.Second),
		shutdown: make(chan bool),
		sessions: sessions,
		config:   config,
		configuration: &webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		},
	}
}

type WebRTCManager struct {
	logger        zerolog.Logger
	settings      webrtc.SettingEngine
	sessions      types.SessionManager
	videoPipeline *gst.Pipeline
	audioPipeline *gst.Pipeline
	cleanup       *time.Ticker
	config        *config.WebRTC
	shutdown      chan bool
	configuration *webrtc.Configuration
}

func (m *WebRTCManager) Start() {
	hid.Display(m.config.Display)

	videoPipeline, err := gst.CreatePipeline(
		m.config.VideoCodec,
		fmt.Sprintf("ximagesrc xid=%s show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue", m.config.Display),
	)

	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
	}

	audioPipeline, err := gst.CreatePipeline(
		m.config.AudioCodec,
		fmt.Sprintf("pulsesrc device=%s ! audioconvert", m.config.Device),
	)

	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
	}

	m.videoPipeline = videoPipeline
	m.audioPipeline = audioPipeline

	videoPipeline.Start()
	audioPipeline.Start()

	go func() {
		defer func() {
			m.logger.Info().Msg("shutdown")
		}()

		for {
			select {
			case <-m.shutdown:
				return
			case sample := <-videoPipeline.Sample:
				if err := m.sessions.WriteVideoSample(sample); err != nil {
					m.logger.Warn().Err(err).Msg("video pipeline failed to write")
				}
			case sample := <-audioPipeline.Sample:
				if err := m.sessions.WriteAudioSample(sample); err != nil {
					m.logger.Warn().Err(err).Msg("audio pipeline failed to write")
				}
			case <-m.cleanup.C:
				hid.Check(time.Second * 10)
			}
		}
	}()

	m.sessions.OnHostCleared(func(id string) {
		hid.Reset()
	})

	m.sessions.OnCreated(func(id string, session types.Session) {
		m.logger.Debug().Str("id", id).Msg("session created")
	})

	m.sessions.OnDestroy(func(id string) {
		m.logger.Debug().Str("id", id).Msg("session destroyed")
	})

	// TODO: log resolution, bit rate and codec parameters
	m.logger.Info().
		Str("video_display", m.config.Display).
		Str("video_codec", m.config.VideoCodec).
		Str("audio_device", m.config.Device).
		Str("audio_codec", m.config.AudioCodec).
		Msgf("webrtc streaming")
}

func (m *WebRTCManager) Shutdown() error {
	m.logger.Info().Msgf("webrtc shutting down")
	m.videoPipeline.Stop()
	m.audioPipeline.Stop()
	m.cleanup.Stop()
	m.shutdown <- true
	return nil
}

func (m *WebRTCManager) CreatePeer(id string, sdp string) (string, types.Peer, error) {
	// Create SessionDescription
	description := webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	}

	// Create MediaEngine based off sdp
	engine := webrtc.MediaEngine{}
	engine.PopulateFromSDP(description)

	// Create API with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(m.settings))

	// Create new peer connection
	connection, err := api.NewPeerConnection(*m.configuration)
	if err != nil {
		return "", nil, err
	}

	// Create video track
	video, err := m.createVideoTrack(engine)
	if err != nil {
		return "", nil, err
	}

	_, err = connection.AddTransceiverFromTrack(video, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})

	if err != nil {
		return "", nil, err
	}

	// Create audio track
	audio, err := m.createAudioTrack(engine)
	if err != nil {
		return "", nil, err
	}

	_, err = connection.AddTransceiverFromTrack(audio, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})

	if err != nil {
		return "", nil, err
	}

	// set remote description
	connection.SetRemoteDescription(description)

	answer, err := connection.CreateAnswer(nil)
	if err != nil {
		return "", nil, err
	}

	if err = connection.SetLocalDescription(answer); err != nil {
		return "", nil, err
	}

	connection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if err = m.handle(id, msg); err != nil {
				m.logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
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

	return answer.SDP, &Peer{
		id:         id,
		api:        api,
		engine:     engine,
		video:      video,
		audio:      audio,
		connection: connection,
	}, nil
}
