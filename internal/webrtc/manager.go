package webrtc

import (
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
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
	videoTrack *webrtc.Track
	audioTrack *webrtc.Track
	videoCodec *webrtc.RTPCodec
	audioCodec *webrtc.RTPCodec
	desktop    types.DesktopManager
	capture    types.CaptureManager
	config     *config.WebRTC
}

func (manager *WebRTCManagerCtx) Start() {
	var err error

	manager.audioTrack, manager.audioCodec, err = manager.createTrack(manager.capture.AudioCodec())
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio track")
	}

	manager.capture.OnAudioFrame(func(sample types.Sample) {
		if err := manager.audioTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("audio pipeline failed to write")
		}
	})

	manager.videoTrack, manager.videoCodec, err = manager.createTrack(manager.capture.VideoCodec())
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

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session) (string, bool, []string, error) {
	configuration := &webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: manager.config.ICEServers,
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	settings := webrtc.SettingEngine{
		LoggerFactory: loggerFactory{
			logger: manager.logger,
		},
	}

	if manager.config.ICELite {
		configuration = &webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		}
		settings.SetLite(true)
	}

	err := settings.SetEphemeralUDPPortRange(manager.config.EphemeralMin, manager.config.EphemeralMax)
	if err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	settings.SetNAT1To1IPs(manager.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)

	// Create MediaEngine based off sdp
	engine := webrtc.MediaEngine{}

	engine.RegisterCodec(manager.audioCodec)
	engine.RegisterCodec(manager.videoCodec)

	// Create API with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(settings))

	// Create new peer connection
	connection, err := api.NewPeerConnection(*configuration)
	if err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	if _, err = connection.AddTransceiverFromTrack(manager.videoTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	if _, err = connection.AddTransceiverFromTrack(manager.audioTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	description, err := connection.CreateOffer(nil)
	if err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	connection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if !session.IsHost() {
				return
			}

			if err = manager.handle(msg); err != nil {
				manager.logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	if err := connection.SetLocalDescription(description); err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateDisconnected:
			manager.logger.Info().Str("id", session.ID()).Msg("peer disconnected successfully")
			session.SetWebRTCConnected(false)
		case webrtc.PeerConnectionStateFailed:
			manager.logger.Warn().Str("id", session.ID()).Msg("peer disconnected with error")
			session.SetWebRTCConnected(false)
		case webrtc.PeerConnectionStateConnected:
			manager.logger.Info().Str("id", session.ID()).Msg("peer connected")
			session.SetWebRTCConnected(true)
		}
	})

	session.SetWebRTCPeer(&WebRTCPeerCtx{
		api:           api,
		engine:        &engine,
		settings:      &settings,
		connection:    connection,
		configuration: configuration,
	})

	return description.SDP, manager.config.ICELite, manager.config.ICEServers, nil
}

func (m *WebRTCManagerCtx) createTrack(codecName string) (*webrtc.Track, *webrtc.RTPCodec, error) {
	var codec *webrtc.RTPCodec
	switch codecName {
	case webrtc.VP8:
		codec = webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)
	case webrtc.VP9:
		codec = webrtc.NewRTPVP9Codec(webrtc.DefaultPayloadTypeVP9, 90000)
	case webrtc.H264:
		codec = webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000)
	case webrtc.Opus:
		codec = webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
	case webrtc.G722:
		codec = webrtc.NewRTPG722Codec(webrtc.DefaultPayloadTypeG722, 8000)
	case webrtc.PCMU:
		codec = webrtc.NewRTPPCMUCodec(webrtc.DefaultPayloadTypePCMU, 8000)
	case webrtc.PCMA:
		codec = webrtc.NewRTPPCMACodec(webrtc.DefaultPayloadTypePCMA, 8000)
	default:
		return nil, nil, fmt.Errorf("unknown codec %s", codecName)
	}

	track, err := webrtc.NewTrack(codec.PayloadType, rand.Uint32(), "stream", "stream", codec)
	if err != nil {
		return nil, nil, err
	}

	return track, codec, nil
}
