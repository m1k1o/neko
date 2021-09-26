package webrtc

import (
	"demodesk/neko/internal/types/codec"
	"demodesk/neko/internal/webrtc/pionlog"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"
)

func (manager *WebRTCManagerCtx) newPeerConnection(codecs []codec.RTPCodec, logger zerolog.Logger) (*webrtc.PeerConnection, error) {
	// create media engine
	engine := &webrtc.MediaEngine{}
	for _, codec := range codecs {
		if err := codec.Register(engine); err != nil {
			return nil, err
		}
	}

	// create setting engine
	settings := webrtc.SettingEngine{
		LoggerFactory: pionlog.New(logger),
	}

	_ = settings.SetEphemeralUDPPortRange(manager.config.EphemeralMin, manager.config.EphemeralMax)
	settings.SetICETimeouts(disconnectedTimeout, failedTimeout, keepAliveInterval)
	settings.SetNAT1To1IPs(manager.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	settings.SetLite(manager.config.ICELite)

	// create interceptor registry
	registry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(engine, registry); err != nil {
		return nil, err
	}

	// create new API
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(engine),
		webrtc.WithSettingEngine(settings),
		webrtc.WithInterceptorRegistry(registry),
	)

	// create new peer connection
	configuration := manager.peerConfiguration()
	return api.NewPeerConnection(configuration)
}

func (manager *WebRTCManagerCtx) peerConfiguration() webrtc.Configuration {
	if manager.config.ICELite {
		return webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		}
	}

	ICEServers := []webrtc.ICEServer{}
	for _, server := range manager.config.ICEServers {
		var credential interface{}
		if server.Credential != "" {
			credential = server.Credential
		} else {
			credential = false
		}

		ICEServers = append(ICEServers, webrtc.ICEServer{
			URLs:       server.URLs,
			Username:   server.Username,
			Credential: credential,
		})
	}

	return webrtc.Configuration{
		ICEServers:   ICEServers,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}
}
