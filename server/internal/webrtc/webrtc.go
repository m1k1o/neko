package webrtc

import (
	"fmt"
	"strings"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/gst"
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/config"
	"n.eko.moe/neko/internal/xorg"
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
	xorg.Display(m.config.Display)

	if !xorg.ValidScreenSize(m.config.ScreenWidth, m.config.ScreenHeight, m.config.ScreenRate) {
		m.logger.Warn().Msgf("invalid screen option %dx%d@%d", m.config.ScreenWidth, m.config.ScreenHeight, m.config.ScreenRate)
	} else {
		if err := xorg.ChangeScreenSize(m.config.ScreenWidth, m.config.ScreenHeight, m.config.ScreenRate); err != nil {
			m.logger.Warn().Err(err).Msg("unable to change screen size")
		}
	}

	videoPipeline, err := gst.CreatePipeline(
		m.config.VideoCodec,
		m.config.Display,
		m.config.VideoParams,
	)

	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
	}

	audioPipeline, err := gst.CreatePipeline(
		m.config.AudioCodec,
		m.config.Device,
		m.config.AudioParams,
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
			case sample := <-m.videoPipeline.Sample:
				if err := m.sessions.WriteVideoSample(sample); err != nil {
					m.logger.Warn().Err(err).Msg("video pipeline failed to write")
				}
			case sample := <-m.audioPipeline.Sample:
				if err := m.sessions.WriteAudioSample(sample); err != nil {
					m.logger.Warn().Err(err).Msg("audio pipeline failed to write")
				}
			case <-m.cleanup.C:
				xorg.CheckKeys(time.Second * 10)
			}
		}
	}()

	m.sessions.OnHostCleared(func(id string) {
		xorg.ResetKeys()
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
		Str("ephemeral_port_range", fmt.Sprintf("%d-%d", m.config.EphemeralMin, m.config.EphemeralMax)).
		Str("nat_ips", strings.Join(m.config.NAT1To1IPs, ",")).
		Str("audio_pipeline_src", audioPipeline.Src).
		Str("video_pipeline_src", videoPipeline.Src).
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

	// Set remote description
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

func (m *WebRTCManager) ChangeScreenSize(width int, height int, rate int) error {
	if !xorg.ValidScreenSize(width, height, rate) {
		return fmt.Errorf("unknown configuration")
	}

	m.videoPipeline.Stop()
	defer func() {
		m.videoPipeline.Start()
		m.logger.Info().Msg("starting pipeline")
	}()

	if err := xorg.ChangeScreenSize(width, height, rate); err != nil {
		return err
	}

	videoPipeline, err := gst.CreatePipeline(
		m.config.VideoCodec,
		m.config.Display,
		m.config.VideoParams,
	)

	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to create new video pipeline")
	}

	m.videoPipeline = videoPipeline

	return nil
}
