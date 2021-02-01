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
	videoCodec codec.RTP
	audioCodec codec.RTP
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
	engine := &webrtc.MediaEngine{}

	if err := manager.videoCodec.Register(engine); err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	if err := manager.audioCodec.Register(engine); err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	// Create API with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(engine), webrtc.WithSettingEngine(settings))

	// Create new peer connection
	connection, err := api.NewPeerConnection(*configuration)
	if err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	_, err = connection.CreateDataChannel("data", nil)
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

	offer, err := connection.CreateOffer(nil)
	if err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	connection.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnMessage(func(message webrtc.DataChannelMessage) {
			if !session.IsHost() {
				return
			}

			if err = manager.handle(message); err != nil {
				manager.logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	// TODO: Refactor, send request to client.
	gatherComplete := webrtc.GatheringCompletePromise(connection)

	if err := connection.SetLocalDescription(offer); err != nil {
		return "", manager.config.ICELite, manager.config.ICEServers, err
	}

	<-gatherComplete

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			manager.logger.Info().Str("id", session.ID()).Msg("peer connected")
			session.SetWebRTCConnected(true)
		case webrtc.PeerConnectionStateDisconnected:
			manager.logger.Info().Str("id", session.ID()).Msg("peer disconnected")
			session.SetWebRTCConnected(false)
		case webrtc.PeerConnectionStateFailed:
			manager.logger.Warn().Str("id", session.ID()).Msg("peer failed")
			session.SetWebRTCConnected(false)
		case webrtc.PeerConnectionStateClosed:
			manager.logger.Warn().Str("id", session.ID()).Msg("peer closed")
			session.SetWebRTCConnected(false)
		}
	})

	session.SetWebRTCPeer(&WebRTCPeerCtx{
		api:           api,
		engine:        engine,
		settings:      &settings,
		connection:    connection,
		configuration: configuration,
	})

	return connection.LocalDescription().SDP, manager.config.ICELite, manager.config.ICEServers, nil
}
