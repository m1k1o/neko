package webrtc

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/config"
)

func New(sessions types.SessionManager, remote types.RemoteManager, config *config.WebRTC) *WebRTCManager {
	return &WebRTCManager{
		logger:   log.With().Str("module", "webrtc").Logger(),
		remote:   remote,
		sessions: sessions,
		config:   config,
	}
}

type WebRTCManager struct {
	logger     zerolog.Logger
	videoTrack *webrtc.TrackLocalStaticSample
	audioTrack *webrtc.TrackLocalStaticSample
	videoCodec webrtc.RTPCodecParameters
	audioCodec webrtc.RTPCodecParameters
	sessions   types.SessionManager
	remote     types.RemoteManager
	config     *config.WebRTC
}

func (manager *WebRTCManager) Start() {
	var err error
	manager.audioTrack, manager.audioCodec, err = manager.createTrack(manager.remote.AudioCodec())
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio track")
	}

	manager.remote.OnAudioFrame(func(sample types.Sample) {
		if err := manager.audioTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("audio pipeline failed to write")
		}
	})

	manager.videoTrack, manager.videoCodec, err = manager.createTrack(manager.remote.VideoCodec())
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create video track")
	}

	manager.remote.OnVideoFrame(func(sample types.Sample) {
		if err := manager.videoTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("video pipeline failed to write")
		}
	})

	manager.logger.Info().
		Str("ice_lite", fmt.Sprintf("%t", manager.config.ICELite)).
		Str("ice_servers", fmt.Sprintf("%+v", manager.config.ICEServers)).
		Str("ephemeral_port_range", fmt.Sprintf("%d-%d", manager.config.EphemeralMin, manager.config.EphemeralMax)).
		Str("nat_ips", strings.Join(manager.config.NAT1To1IPs, ",")).
		Msgf("webrtc starting")
}

func (manager *WebRTCManager) Shutdown() error {
	manager.logger.Info().Msgf("webrtc shutting down")
	return nil
}

func (manager *WebRTCManager) CreatePeer(id string, session types.Session) (types.Peer, error) {
	configuration := &webrtc.Configuration{
		ICEServers:   manager.config.ICEServers,
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

	_ = settings.SetEphemeralUDPPortRange(manager.config.EphemeralMin, manager.config.EphemeralMax)
	settings.SetNAT1To1IPs(manager.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	settings.SetICETimeouts(6*time.Second, 6*time.Second, 3*time.Second)
	settings.SetSRTPReplayProtectionWindow(512)

	// Create MediaEngine based off sdp
	engine := webrtc.MediaEngine{}

	_ = engine.RegisterCodec(manager.audioCodec, webrtc.RTPCodecTypeAudio)
	_ = engine.RegisterCodec(manager.videoCodec, webrtc.RTPCodecTypeVideo)

	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(&engine, i); err != nil {
		return nil, err
	}

	// Create API with MediaEngine and SettingEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&engine), webrtc.WithSettingEngine(settings), webrtc.WithInterceptorRegistry(i))

	// Create new peer connection
	connection, err := api.NewPeerConnection(*configuration)
	if err != nil {
		return nil, err
	}

	negotiated := true
	_, err = connection.CreateDataChannel("data", &webrtc.DataChannelInit{
		Negotiated: &negotiated,
	})
	if err != nil {
		return nil, err
	}

	connection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if err = manager.handle(id, msg); err != nil {
				manager.logger.Warn().Err(err).Msg("data handle failed")
			}
		})
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	connection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		manager.logger.Info().
			Str("connection_state", connectionState.String()).
			Msg("connection state has changed")
	})

	rtpVideo, err := connection.AddTrack(manager.videoTrack)
	if err != nil {
		return nil, err
	}

	rtpAudio, err := connection.AddTrack(manager.audioTrack)
	if err != nil {
		return nil, err
	}

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateDisconnected:
			manager.logger.Info().Str("id", id).Msg("peer disconnected")
			manager.sessions.Destroy(id)
		case webrtc.PeerConnectionStateFailed:
			manager.logger.Warn().Str("id", id).Msg("peer failed")
			manager.sessions.Destroy(id)
		case webrtc.PeerConnectionStateClosed:
			manager.logger.Info().Str("id", id).Msg("peer closed")
			manager.sessions.Destroy(id)
		case webrtc.PeerConnectionStateConnected:
			manager.logger.Info().Str("id", id).Msg("peer connected")
			if err = session.SetConnected(true); err != nil {
				manager.logger.Warn().Err(err).Msg("unable to set connected on peer")
				manager.sessions.Destroy(id)
			}
		}
	})

	peer := &Peer{
		id:            id,
		api:           api,
		engine:        &engine,
		manager:       manager,
		settings:      &settings,
		connection:    connection,
		configuration: configuration,
	}

	connection.OnNegotiationNeeded(func() {
		manager.logger.Warn().Msg("negotiation is needed")

		sdp, err := peer.CreateOffer()
		if err != nil {
			manager.logger.Err(err).Msg("creating offer failed")
			return
		}

		err = session.SignalLocalOffer(sdp)
		if err != nil {
			manager.logger.Warn().Err(err).Msg("sending SignalLocalOffer failed")
			return
		}
	})

	connection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			manager.logger.Info().Msg("sent all ICECandidates")
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			manager.logger.Warn().Err(err).Msg("converting ICECandidate to json failed")
			return
		}

		if err := session.SignalCandidate(string(candidateString)); err != nil {
			manager.logger.Warn().Err(err).Msg("sending SignalCandidate failed")
			return
		}
	})

	if err := session.SetPeer(peer); err != nil {
		return nil, err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpVideo.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpAudio.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	return peer, nil
}

func (manager *WebRTCManager) ICELite() bool {
	return manager.config.ICELite
}

func (manager *WebRTCManager) ICEServers() []webrtc.ICEServer {
	return manager.config.ICEServers
}

func (manager *WebRTCManager) createTrack(codecName string) (*webrtc.TrackLocalStaticSample, webrtc.RTPCodecParameters, error) {
	var codec webrtc.RTPCodecParameters

	id := ""
	fb := []webrtc.RTCPFeedback{}

	switch codecName {
	case "VP8":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: fb}, PayloadType: 96}
		id = "video"
	case "VP9":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP9, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: fb}, PayloadType: 98}
		id = "video"
	case "H264":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000, Channels: 0, SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f", RTCPFeedback: fb}, PayloadType: 102}
		id = "video"
	case "Opus":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "", RTCPFeedback: fb}, PayloadType: 111}
		id = "audio"
	case "G722":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeG722, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: fb}, PayloadType: 9}
		id = "audio"
	case "PCMU":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypePCMU, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: fb}, PayloadType: 0}
		id = "audio"
	case "PCMA":
		codec = webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypePCMA, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: fb}, PayloadType: 8}
		id = "audio"
	default:
		return nil, codec, fmt.Errorf("unknown codec %s", codecName)
	}

	track, err := webrtc.NewTrackLocalStaticSample(codec.RTPCodecCapability, id, "stream")
	if err != nil {
		return nil, codec, err
	}

	return track, codec, nil
}
