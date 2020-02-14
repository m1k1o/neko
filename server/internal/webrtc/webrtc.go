package webrtc

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
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
	settings.SetEphemeralUDPPortRange(config.EphemeralMin, config.EphemeralMax)
	settings.SetNAT1To1IPs(config.NAT1To1IPs, webrtc.ICECandidateTypeHost)

	// Create MediaEngine based off sdp
	engine := webrtc.MediaEngine{}
	engine.RegisterDefaultCodecs()

	// Create API with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(settings))

	return &WebRTCManager{
		logger:   logger,
		settings: settings,
		cleanup:  time.NewTicker(1 * time.Second),
		shutdown: make(chan bool),
		sessions: sessions,
		engine:   engine,
		config:   config,
		api:      api,
		configuration: &webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		},
	}
}

type WebRTCManager struct {
	logger        zerolog.Logger
	settings      webrtc.SettingEngine
	engine        webrtc.MediaEngine
	api           *webrtc.API
	videoTrack    *webrtc.Track
	audioTrack    *webrtc.Track
	videoPipeline *gst.Pipeline
	audioPipeline *gst.Pipeline
	sessions      types.SessionManager
	cleanup       *time.Ticker
	config        *config.WebRTC
	shutdown      chan bool
	configuration *webrtc.Configuration
}

func (m *WebRTCManager) Start() {
	// Set display and change to default resolution
	xorg.Display(m.config.Display)
	if !xorg.ValidScreenSize(m.config.ScreenWidth, m.config.ScreenHeight, m.config.ScreenRate) {
		m.logger.Warn().Msgf("invalid screen option %dx%d@%d", m.config.ScreenWidth, m.config.ScreenHeight, m.config.ScreenRate)
	} else {
		if err := xorg.ChangeScreenSize(m.config.ScreenWidth, m.config.ScreenHeight, m.config.ScreenRate); err != nil {
			m.logger.Warn().Err(err).Msg("unable to change screen size")
		}
	}

	var err error
	m.videoPipeline, m.videoTrack, err = m.createTrack(m.config.VideoCodec, m.config.Display, m.config.VideoParams)
	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
	}

	m.audioPipeline, m.audioTrack, err = m.createTrack(m.config.AudioCodec, m.config.Device, m.config.AudioParams)
	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
	}

	go func() {
		defer func() {
			m.logger.Info().Msg("shutdown")
		}()

		for {
			select {
			case <-m.shutdown:
				return
			case sample := <-m.videoPipeline.Sample:
				if err := m.videoTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
					m.logger.Warn().Err(err).Msg("video pipeline failed to write")
				}
			case sample := <-m.audioPipeline.Sample:
				if err := m.audioTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
					m.logger.Warn().Err(err).Msg("audio pipeline failed to write")
				}
			case <-m.cleanup.C:
				xorg.CheckKeys(time.Second * 10)
			}
		}
	}()

	m.videoPipeline.Start()
	m.audioPipeline.Start()

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
		Str("audio_pipeline_src", m.audioPipeline.Src).
		Str("video_pipeline_src", m.videoPipeline.Src).
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

func (m *WebRTCManager) CreatePeer(id string, session types.Session) (string, error) {
	// Create new peer connection
	connection, err := m.api.NewPeerConnection(*m.configuration)
	if err != nil {
		return "", err
	}

	if _, err = connection.AddTransceiverFromTrack(m.videoTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return "", err
	}

	if _, err = connection.AddTransceiverFromTrack(m.audioTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return "", err
	}

	description, err := connection.CreateOffer(nil)
	if err != nil {
		return "", err
	}

	connection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if err = m.handle(id, msg); err != nil {
				m.logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	connection.SetLocalDescription(description)
	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateDisconnected:
		case webrtc.PeerConnectionStateFailed:
			m.logger.Info().Str("id", id).Msg("peer disconnected")
			m.sessions.Destroy(id)
			break
		case webrtc.PeerConnectionStateConnected:
			m.logger.Info().Str("id", id).Msg("peer connected")
			if err = session.SetConnected(true); err != nil {
				m.logger.Warn().Err(err).Msg("unable to set connected on peer")
				m.sessions.Destroy(id)
			}
			break
		}
	})

	if err := session.SetPeer(&Peer{
		id:         id,
		manager:    m,
		connection: connection,
	}); err != nil {
		return "", err
	}

	return description.SDP, nil
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
