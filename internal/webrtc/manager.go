package webrtc

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/internal/webrtc/cursor"
	"github.com/demodesk/neko/internal/webrtc/pionlog"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

// the duration without network activity before a Agent is considered disconnected. Default is 5 Seconds
const disconnectedTimeout = 4 * time.Second

// the duration without network activity before a Agent is considered failed after disconnected. Default is 25 Seconds
const failedTimeout = 6 * time.Second

// how often the ICE Agent sends extra traffic if there is no activity, if media is flowing no traffic will be sent. Default is 2 seconds
const keepAliveInterval = 2 * time.Second

// send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
const rtcpPLIInterval = 3 * time.Second

func New(desktop types.DesktopManager, capture types.CaptureManager, config *config.WebRTC) *WebRTCManagerCtx {
	configuration := webrtc.Configuration{
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	if !config.ICELite {
		ICEServers := []webrtc.ICEServer{}
		for _, server := range config.ICEServers {
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

		configuration.ICEServers = ICEServers
	}

	return &WebRTCManagerCtx{
		logger:  log.With().Str("module", "webrtc").Logger(),
		config:  config,
		metrics: newMetrics(),

		webrtcConfiguration: configuration,

		desktop:     desktop,
		capture:     capture,
		curImage:    cursor.NewImage(desktop),
		curPosition: cursor.NewPosition(),
	}
}

type WebRTCManagerCtx struct {
	logger  zerolog.Logger
	config  *config.WebRTC
	metrics *metricsCtx
	peerId  int32

	desktop     types.DesktopManager
	capture     types.CaptureManager
	curImage    *cursor.ImageCtx
	curPosition *cursor.PositionCtx

	webrtcConfiguration webrtc.Configuration

	tcpMux ice.TCPMux
	udpMux ice.UDPMux

	camStop, micStop *func()
}

func (manager *WebRTCManagerCtx) Start() {
	manager.curImage.Start()

	logger := pionlog.New(manager.logger)

	// add TCP Mux listener
	if manager.config.TCPMux > 0 {
		tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
			IP:   net.IP{0, 0, 0, 0},
			Port: manager.config.TCPMux,
		})

		if err != nil {
			manager.logger.Panic().Err(err).Msg("unable to setup ice TCP mux")
		}

		manager.tcpMux = webrtc.NewICETCPMux(logger.NewLogger("ice-tcp"), tcpListener, 32)
	}

	// add UDP Mux listener
	if manager.config.UDPMux > 0 {
		udpListener, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.IP{0, 0, 0, 0},
			Port: manager.config.UDPMux,
		})

		if err != nil {
			manager.logger.Panic().Err(err).Msg("unable to setup ice UDP mux")
		}

		manager.udpMux = webrtc.NewICEUDPMux(logger.NewLogger("ice-udp"), udpListener)
	}

	manager.logger.Info().
		Bool("icelite", manager.config.ICELite).
		Bool("icetrickle", manager.config.ICETrickle).
		Interface("iceservers", manager.config.ICEServers).
		Str("nat1to1", strings.Join(manager.config.NAT1To1IPs, ",")).
		Str("epr", fmt.Sprintf("%d-%d", manager.config.EphemeralMin, manager.config.EphemeralMax)).
		Int("tcpmux", manager.config.TCPMux).
		Int("udpmux", manager.config.UDPMux).
		Msg("webrtc starting")
}

func (manager *WebRTCManagerCtx) Shutdown() error {
	manager.logger.Info().Msg("shutdown")

	manager.curImage.Shutdown()
	manager.curPosition.Shutdown()

	return nil
}

func (manager *WebRTCManagerCtx) ICEServers() []types.ICEServer {
	return manager.config.ICEServers
}

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
	return api.NewPeerConnection(manager.webrtcConfiguration)
}

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session, videoID string) (*webrtc.SessionDescription, error) {
	id := atomic.AddInt32(&manager.peerId, 1)
	manager.metrics.NewConnection(session)

	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Int32("peer_id", id).Logger()
	logger.Info().Msg("creating webrtc peer")

	// all audios must have the same codec
	audio := manager.capture.Audio()
	audioCodec := audio.Codec()

	// all videos must have the same codec
	video := manager.capture.Video()
	videoCodec := video.Codec()

	connection, err := manager.newPeerConnection([]codec.RTPCodec{
		audioCodec,
		videoCodec,
	}, logger)
	if err != nil {
		return nil, err
	}

	// asynchronously send local ICE Candidates
	if manager.config.ICETrickle {
		connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate == nil {
				logger.Debug().Msg("all local ice candidates sent")
				return
			}

			session.Send(
				event.SIGNAL_CANDIDATE,
				message.SignalCandidate{
					ICECandidateInit: candidate.ToJSON(),
				})
		})
	}

	// audio track

	audioTrack, err := NewTrack(logger, audioCodec, connection)
	if err != nil {
		return nil, err
	}

	// set stream for audio track
	err = audioTrack.SetStream(audio)
	if err != nil {
		return nil, err
	}

	// video track

	videoTrack, err := NewTrack(logger, videoCodec, connection)
	if err != nil {
		return nil, err
	}

	// let video stream bucket manager handle stream subscriptions
	err = video.SetReceiver(videoTrack)
	if err != nil {
		return nil, err
	}

	// set default video id
	err = videoTrack.SetVideoID(videoID)
	if err != nil {
		return nil, err
	}
	manager.metrics.SetVideoID(session, videoID)

	// data channel

	dataChannel, err := connection.CreateDataChannel("data", nil)
	if err != nil {
		return nil, err
	}

	peer := &WebRTCPeerCtx{
		logger:      logger,
		connection:  connection,
		dataChannel: dataChannel,
		changeVideo: func(videoID string) error {
			if err := videoTrack.SetVideoID(videoID); err != nil {
				return err
			}

			manager.metrics.SetVideoID(session, videoID)
			return nil
		},
		setPaused: func(isPaused bool) {
			videoTrack.SetPaused(isPaused)
			audioTrack.SetPaused(isPaused)
		},
		iceTrickle: manager.config.ICETrickle,
	}

	connection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		logger := logger.With().
			Str("kind", track.Kind().String()).
			Str("mime", track.Codec().RTPCodecCapability.MimeType).
			Logger()

		logger.Info().Msgf("received new remote track")

		if !session.Profile().CanShareMedia {
			logger.Warn().Msg("media sharing is disabled for this session")
			receiver.Stop()
			return
		}

		// parse codec from remote track
		codec, ok := codec.ParseRTC(track.Codec())
		if !ok {
			logger.Warn().Msg("remote track with unknown codec")
			receiver.Stop()
			return
		}

		var srcManager types.StreamSrcManager

		stopped := false
		stopFn := func() {
			if stopped {
				return
			}

			stopped = true
			receiver.Stop()
			srcManager.Stop()
			logger.Info().Msg("remote track stopped")
		}

		if track.Kind() == webrtc.RTPCodecTypeAudio {
			// audio -> microphone
			srcManager = manager.capture.Microphone()
			defer stopFn()

			if manager.micStop != nil {
				(*manager.micStop)()
			}
			manager.micStop = &stopFn
		} else if track.Kind() == webrtc.RTPCodecTypeVideo {
			// video -> webcam
			srcManager = manager.capture.Webcam()
			defer stopFn()

			if manager.camStop != nil {
				(*manager.camStop)()
			}
			manager.camStop = &stopFn
		} else {
			logger.Warn().Msg("remote track with unsupported codec type")
			receiver.Stop()
			return
		}

		err := srcManager.Start(codec)
		if err != nil {
			logger.Err(err).Msg("failed to start pipeline")
			return
		}

		ticker := time.NewTicker(rtcpPLIInterval)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				err := connection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if err != nil {
					logger.Err(err).Msg("remote track rtcp send err")
				}
			}
		}()

		buf := make([]byte, 1400)
		for {
			i, _, err := track.Read(buf)
			if err != nil {
				logger.Warn().Err(err).Msg("failed read from remote track")
				break
			}

			srcManager.Push(buf[:i])
		}

		logger.Info().Msg("remote track data finished")
	})

	connection.OnDataChannel(func(dc *webrtc.DataChannel) {
		logger.Info().Interface("data_channel", dc).Msg("got remote data channel")
	})

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			session.SetWebRTCConnected(peer, true)
		case webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateFailed:
			connection.Close()
		case webrtc.PeerConnectionStateClosed:
			session.SetWebRTCConnected(peer, false)
			video.RemoveReceiver(videoTrack)
			audioTrack.RemoveStream()
		}

		manager.metrics.SetState(session, state)
	})

	cursorImage := func(entry *cursor.ImageEntry) {
		if err := peer.SendCursorImage(entry.Cursor, entry.Image); err != nil {
			logger.Err(err).Msg("could not send cursor image")
		}
	}

	cursorPosition := func(x, y int) {
		if session.IsHost() {
			return
		}

		if err := peer.SendCursorPosition(x, y); err != nil {
			logger.Err(err).Msg("could not send cursor position")
		}
	}

	dataChannel.OnOpen(func() {
		manager.curImage.AddListener(&cursorImage)
		manager.curPosition.AddListener(&cursorPosition)

		// send initial cursor image
		entry, err := manager.curImage.Get()
		if err == nil {
			cursorImage(entry)
		} else {
			logger.Err(err).Msg("failed to get cursor image")
		}

		// send initial cursor position
		x, y := manager.desktop.GetCursorPosition()
		cursorPosition(x, y)
	})

	dataChannel.OnClose(func() {
		manager.curImage.RemoveListener(&cursorImage)
		manager.curPosition.RemoveListener(&cursorPosition)
	})

	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
		if err := manager.handle(message.Data, session); err != nil {
			logger.Err(err).Msg("data handle failed")
		}
	})

	session.SetWebRTCPeer(peer)

	offer, err := peer.CreateOffer(false)
	if err != nil {
		return nil, err
	}

	// on negotiation needed handler must be registered after creating initial
	// offer, otherwise it can fire and intercept sucessful negotiation

	connection.OnNegotiationNeeded(func() {
		logger.Warn().Msg("negotiation is needed")

		if connection.SignalingState() != webrtc.SignalingStateStable {
			logger.Warn().Msg("connection isn't stable yet; postponing...")
			return
		}

		offer, err := peer.CreateOffer(false)
		if err != nil {
			logger.Err(err).Msg("sdp offer failed")
			return
		}

		session.Send(
			event.SIGNAL_OFFER,
			message.SignalDescription{
				SDP: offer.SDP,
			})
	})

	videoTrack.OnRTCP(func(p rtcp.Packet) {
		switch rtcpPacket := p.(type) {
		case *rtcp.ReceiverEstimatedMaximumBitrate: // TODO: Deprecated.
			manager.metrics.SetReceiverEstimatedMaximumBitrate(session, rtcpPacket.Bitrate)

		case *rtcp.ReceiverReport:
			l := len(rtcpPacket.Reports)
			if l > 0 {
				// use only last report
				manager.metrics.SetReceiverReport(session, rtcpPacket.Reports[l-1])
			}
		}
	})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				break
			}

			stats := connection.GetStats()
			data, ok := stats["iceTransport"].(webrtc.TransportStats)
			if ok {
				manager.metrics.SetIceTransportStats(session, data)
			}

			data, ok = stats["sctpTransport"].(webrtc.TransportStats)
			if ok {
				manager.metrics.SetSctpTransportStats(session, data)
			}

			for _, entry := range stats {
				// only remote ice candidate stats
				candidate, ok := entry.(webrtc.ICECandidateStats)
				if ok && candidate.Type == webrtc.StatsTypeRemoteCandidate {
					manager.metrics.NewICECandidate(session, candidate.ID)
				}
			}
		}
	}()

	return offer, nil
}

func (manager *WebRTCManagerCtx) SetCursorPosition(x, y int) {
	manager.curPosition.Set(x, y)
}
