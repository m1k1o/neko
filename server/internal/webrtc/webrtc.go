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
	return &WebRTCManager{
		logger:   log.With().Str("module", "webrtc").Logger(),
		cleanup:  time.NewTicker(1 * time.Second),
		shutdown: make(chan bool),
		sessions: sessions,
		config:   config,
	}
}

type WebRTCManager struct {
	logger        zerolog.Logger
	videoTrack    *webrtc.Track
	audioTrack    *webrtc.Track
	videoPipeline *gst.Pipeline
	audioPipeline *gst.Pipeline
	videoCodec    *webrtc.RTPCodec
	audioCodec    *webrtc.RTPCodec
	sessions      types.SessionManager
	cleanup       *time.Ticker
	config        *config.WebRTC
	shutdown      chan bool
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
	m.videoPipeline, m.videoTrack, m.videoCodec, err = m.createTrack(m.config.VideoCodec, m.config.Display, m.config.VideoParams)
	if err != nil {
		m.logger.Panic().Err(err).Msg("unable to start webrtc manager")
	}

	m.audioPipeline, m.audioTrack, m.audioCodec, err = m.createTrack(m.config.AudioCodec, m.config.Device, m.config.AudioParams)
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
		Str("audio_pipeline_src", m.audioPipeline.Src).
		Str("video_pipeline_src", m.videoPipeline.Src).
		Str("ice_lite", fmt.Sprintf("%t", m.config.ICELite)).
		Str("ice_servers", strings.Join(m.config.ICEServers, ",")).
		Str("ephemeral_port_range", fmt.Sprintf("%d-%d", m.config.EphemeralMin, m.config.EphemeralMax)).
		Str("nat_ips", strings.Join(m.config.NAT1To1IPs, ",")).
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

func (m *WebRTCManager) CreatePeer(id string, session types.Session) (string, bool, []string, error) {
	configuration := &webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: m.config.ICEServers,
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	settings := webrtc.SettingEngine{
		LoggerFactory: loggerFactory{
			logger: m.logger,
		},
	}

	if m.config.ICELite {
		configuration = &webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		}
		settings.SetLite(true)
	}

	settings.SetEphemeralUDPPortRange(m.config.EphemeralMin, m.config.EphemeralMax)
	settings.SetNAT1To1IPs(m.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)

	// Create MediaEngine based off sdp
	engine := webrtc.MediaEngine{}
	// engine.RegisterDefaultCodecs()
	engine.RegisterCodec(m.audioCodec)
	engine.RegisterCodec(m.videoCodec)

	// Create API with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(settings))

	// Create new peer connection
	connection, err := api.NewPeerConnection(*configuration)
	if err != nil {
		return "", m.config.ICELite, m.config.ICEServers, err
	}

	if _, err = connection.AddTransceiverFromTrack(m.videoTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return "", m.config.ICELite, m.config.ICEServers, err
	}

	if _, err = connection.AddTransceiverFromTrack(m.audioTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return "", m.config.ICELite, m.config.ICEServers, err
	}

	description, err := connection.CreateOffer(nil)
	if err != nil {
		return "", m.config.ICELite, m.config.ICEServers, err
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
		id:            id,
		api:           api,
		engine:        &engine,
		manager:       m,
		settings:      &settings,
		connection:    connection,
		configuration: configuration,
	}); err != nil {
		return "", m.config.ICELite, m.config.ICEServers, err
	}

	return description.SDP, m.config.ICELite, m.config.ICEServers, nil
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
