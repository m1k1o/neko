package webrtc

import (
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/internal/webrtc/pionlog"
	"github.com/demodesk/neko/pkg/types/codec"
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

	settings.SetICETimeouts(disconnectedTimeout, failedTimeout, keepAliveInterval)
	settings.SetNAT1To1IPs(manager.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	settings.SetLite(manager.config.ICELite)

	var networkType []webrtc.NetworkType

	// udp candidates
	if manager.udpMux != nil {
		settings.SetICEUDPMux(manager.udpMux)
		networkType = append(networkType,
			webrtc.NetworkTypeUDP4,
			webrtc.NetworkTypeUDP6,
		)
	} else if manager.config.EphemeralMax != 0 {
		_ = settings.SetEphemeralUDPPortRange(manager.config.EphemeralMin, manager.config.EphemeralMax)
		networkType = append(networkType,
			webrtc.NetworkTypeUDP4,
			webrtc.NetworkTypeUDP6,
		)
	}

	// tcp candidates
	if manager.tcpMux != nil {
		settings.SetICETCPMux(manager.tcpMux)
		networkType = append(networkType,
			webrtc.NetworkTypeTCP4,
			webrtc.NetworkTypeTCP6,
		)
	}

	// enable support for TCP and UDP ICE candidates
	settings.SetNetworkTypes(networkType)

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
		var credential any
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
