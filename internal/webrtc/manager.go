package webrtc

import (
	"fmt"
	"io"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/config"
)

func New(desktop types.DesktopManager, capture types.CaptureManager, config *config.WebRTC) *WebRTCManagerCtx {
	return &WebRTCManagerCtx{
		logger:   log.With().Str("module", "webrtc").Logger(),
		desktop:  desktop,
		capture:  capture,
		config:   config,
	}
}

type WebRTCManagerCtx struct {
	logger     zerolog.Logger
	videoTrack *webrtc.TrackLocalStaticSample
	audioTrack *webrtc.TrackLocalStaticSample
	videoCodec codec.RTPCodec
	audioCodec codec.RTPCodec
	desktop    types.DesktopManager
	capture    types.CaptureManager
	config     *config.WebRTC
}

func (manager *WebRTCManagerCtx) Start() {
	var err error

	// create audio track
	manager.audioCodec = manager.capture.AudioCodec()
	manager.audioTrack, err = webrtc.NewTrackLocalStaticSample(manager.audioCodec.Capability, "audio", "stream")
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio track")
	}

	manager.capture.OnAudioFrame(func(sample types.Sample) {
		if err := manager.audioTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("audio pipeline failed to write")
		}
	})

	// create video track
	manager.videoCodec = manager.capture.VideoCodec()
	manager.videoTrack, err = webrtc.NewTrackLocalStaticSample(manager.videoCodec.Capability, "video", "stream")
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create video track")
	}

	manager.capture.OnVideoFrame(func(sample types.Sample) {
		if err := manager.videoTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("video pipeline failed to write")
		}
	})

	manager.logger.Info().
		Str("ice_lite", fmt.Sprintf("%t", manager.config.ICELite)).
		Str("ice_servers", strings.Join(manager.config.ICEServers, ",")).
		Str("ephemeral_port_range", fmt.Sprintf("%d-%d", manager.config.EphemeralMin, manager.config.EphemeralMax)).
		Str("nat_ips", strings.Join(manager.config.NAT1To1IPs, ",")).
		Msgf("webrtc starting")
}

func (manager *WebRTCManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("webrtc shutting down")
	return nil
}

func (manager *WebRTCManagerCtx) ICELite() bool {
	return manager.config.ICELite
}

func (manager *WebRTCManagerCtx) ICEServers() []string {
	return manager.config.ICEServers
}

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session) (*webrtc.SessionDescription, error) {
	logger := manager.logger.With().Str("id", session.ID()).Logger()

	engine, err := manager.mediaEngine()
	if err != nil {
		return nil, err
	}

	// Custom settings & configuration
	settings := manager.apiSettings(logger)
	configuration := manager.apiConfiguration()

	// Create NewAPI with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(*settings))

	connection, err := api.NewPeerConnection(*configuration)
	if err != nil {
		return nil, err
	}

	// Asynchronously send local ICE Candidates
	connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			logger.Debug().Msg("all local ice candidates sent")
			return
		}

		ICECandidateInit := candidate.ToJSON()
		err := session.Send(
			message.SignalCandidate{
				Event:            event.SIGNAL_CANDIDATE,
				ICECandidateInit: &ICECandidateInit,
			})

		if err != nil {
			logger.Warn().Err(err).Msg("sending ice candidate failed")
		}
	})

	if err := manager.registerTracks(connection); err != nil {
		return nil, err
	}

	offer, err := connection.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	if err := connection.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			session.SetWebRTCConnected(true)
		case webrtc.PeerConnectionStateDisconnected:
			fallthrough
		case webrtc.PeerConnectionStateFailed:
			connection.Close()
		case webrtc.PeerConnectionStateClosed:
			session.SetWebRTCConnected(false)
		}
	})

	connection.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnMessage(func(message webrtc.DataChannelMessage) {
			if !session.IsHost() {
				return
			}

			if err = manager.handle(message); err != nil {
				logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	session.SetWebRTCPeer(&WebRTCPeerCtx{
		api:           api,
		engine:        engine,
		settings:      settings,
		connection:    connection,
		configuration: configuration,
	})

	return connection.LocalDescription(), nil
}

func (manager *WebRTCManagerCtx) mediaEngine() (*webrtc.MediaEngine, error) {
	// Create MediaEngine
	engine := &webrtc.MediaEngine{}

	if err := manager.videoCodec.Register(engine); err != nil {
		return nil, err
	}

	if err := manager.audioCodec.Register(engine); err != nil {
		return nil, err
	}

	return engine, nil
}

func (manager *WebRTCManagerCtx) apiSettings(logger zerolog.Logger) *webrtc.SettingEngine {
	settings := &webrtc.SettingEngine{
		LoggerFactory: loggerFactory{
			logger: logger,
		},
	}

	//nolint
	settings.SetEphemeralUDPPortRange(manager.config.EphemeralMin, manager.config.EphemeralMax)
	settings.SetNAT1To1IPs(manager.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	settings.SetLite(manager.config.ICELite)
	
	return settings
}

func (manager *WebRTCManagerCtx) apiConfiguration() *webrtc.Configuration {
	if manager.config.ICELite {
		return &webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		}
	}

	return &webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: manager.config.ICEServers,
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}
}

func (manager *WebRTCManagerCtx) registerTracks(connection *webrtc.PeerConnection) error {
	if _, err := connection.AddTransceiverFromTrack(manager.videoTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return err
	}

	if _, err := connection.AddTransceiverFromTrack(manager.audioTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return err
	}

	_, err :=  connection.CreateDataChannel("data", nil)
	return err
}
