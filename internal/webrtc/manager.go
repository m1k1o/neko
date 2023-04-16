package webrtc

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/interceptor/pkg/gcc"
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

const (
	// size of receiving channel used to buffer incoming TCP packets
	tcpReadChanBufferSize = 50

	// size of buffer used to buffer outgoing TCP packets. Default is 4MB
	tcpWriteBufferSizeInBytes = 4 * 1024 * 1024

	// the duration without network activity before a Agent is considered disconnected. Default is 5 Seconds
	disconnectedTimeout = 4 * time.Second

	// the duration without network activity before a Agent is considered failed after disconnected. Default is 25 Seconds
	failedTimeout = 6 * time.Second

	// how often the ICE Agent sends extra traffic if there is no activity, if media is flowing no traffic will be sent. Default is 2 seconds
	keepAliveInterval = 2 * time.Second

	// send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	rtcpPLIInterval = 3 * time.Second
)

func New(desktop types.DesktopManager, capture types.CaptureManager, config *config.WebRTC) *WebRTCManagerCtx {
	logger := log.With().Str("module", "webrtc").Logger()

	configuration := webrtc.Configuration{
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	if !config.ICELite {
		ICEServers := []webrtc.ICEServer{}
		for _, server := range config.ICEServersBackend {
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
		logger:  logger,
		config:  config,
		metrics: newMetricsManager(),

		webrtcConfiguration: configuration,

		desktop:     desktop,
		capture:     capture,
		curImage:    cursor.NewImage(logger, desktop),
		curPosition: cursor.NewPosition(logger),
	}
}

type WebRTCManagerCtx struct {
	logger  zerolog.Logger
	config  *config.WebRTC
	metrics *metricsManager
	peerId  int32

	desktop     types.DesktopManager
	capture     types.CaptureManager
	curImage    cursor.Image
	curPosition cursor.Position

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
			manager.logger.Fatal().Err(err).Msg("unable to setup ice TCP mux")
		}

		manager.tcpMux = ice.NewTCPMuxDefault(ice.TCPMuxParams{
			Listener:        tcpListener,
			Logger:          logger.NewLogger("ice-tcp"),
			ReadBufferSize:  tcpReadChanBufferSize,
			WriteBufferSize: tcpWriteBufferSizeInBytes,
		})
	}

	// add UDP Mux listener
	if manager.config.UDPMux > 0 {
		var err error
		manager.udpMux, err = ice.NewMultiUDPMuxFromPort(manager.config.UDPMux,
			ice.UDPMuxFromPortWithLogger(logger.NewLogger("ice-udp")),
		)

		if err != nil {
			manager.logger.Fatal().Err(err).Msg("unable to setup ice UDP mux")
		}
	}

	manager.logger.Info().
		Bool("icelite", manager.config.ICELite).
		Bool("icetrickle", manager.config.ICETrickle).
		Interface("iceservers-frontend", manager.config.ICEServersFrontend).
		Interface("iceservers-backend", manager.config.ICEServersBackend).
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
	return manager.config.ICEServersFrontend
}

func (manager *WebRTCManagerCtx) newPeerConnection(logger zerolog.Logger, codecs []codec.RTPCodec, bitrate int) (*webrtc.PeerConnection, cc.BandwidthEstimator, error) {
	// create media engine
	engine := &webrtc.MediaEngine{}
	for _, codec := range codecs {
		if err := codec.Register(engine); err != nil {
			return nil, nil, err
		}
	}

	// create setting engine
	settings := webrtc.SettingEngine{
		LoggerFactory: pionlog.New(logger),
	}

	settings.DisableMediaEngineCopy(true)
	settings.SetICETimeouts(disconnectedTimeout, failedTimeout, keepAliveInterval)
	settings.SetNAT1To1IPs(manager.config.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	settings.SetLite(manager.config.ICELite)
	// make sure server answer sdp setup as passive, to not force DTLS renegotiation
	// otherwise iOS renegotiation fails with: Failed to set SSL role for the transport.
	settings.SetAnsweringDTLSRole(webrtc.DTLSRoleServer)

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

	// create bandwidth estimator
	estimatorChan := make(chan cc.BandwidthEstimator, 1)
	if manager.config.EstimatorEnabled {
		congestionController, err := cc.NewInterceptor(func() (cc.BandwidthEstimator, error) {
			if bitrate == 0 {
				bitrate = manager.config.EstimatorInitialBitrate
			}

			return gcc.NewSendSideBWE(
				gcc.SendSideBWEInitialBitrate(bitrate),
				gcc.SendSideBWEPacer(gcc.NewNoOpPacer()),
			)
		})
		if err != nil {
			return nil, nil, err
		}

		congestionController.OnNewPeerConnection(func(id string, estimator cc.BandwidthEstimator) {
			estimatorChan <- estimator
		})

		registry.Add(congestionController)
		if err = webrtc.ConfigureTWCCHeaderExtensionSender(engine, registry); err != nil {
			return nil, nil, err
		}
	} else {
		// no estimator, send nil
		estimatorChan <- nil
	}

	if err := webrtc.RegisterDefaultInterceptors(engine, registry); err != nil {
		return nil, nil, err
	}

	// create new API
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(engine),
		webrtc.WithSettingEngine(settings),
		webrtc.WithInterceptorRegistry(registry),
	)

	// create new peer connection
	configuration := manager.webrtcConfiguration
	connection, err := api.NewPeerConnection(configuration)
	return connection, <-estimatorChan, err
}

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session, bitrate int, videoAuto bool) (*webrtc.SessionDescription, error) {
	id := atomic.AddInt32(&manager.peerId, 1)

	// get metrics for session
	metrics := manager.metrics.getBySession(session)
	metrics.NewConnection()

	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Int32("peer_id", id).Logger()
	logger.Info().Msg("creating webrtc peer")

	// all audios must have the same codec
	audio := manager.capture.Audio()
	audioCodec := audio.Codec()

	// all videos must have the same codec
	video := manager.capture.Video()
	videoCodec := video.Codec()

	connection, estimator, err := manager.newPeerConnection(logger, []codec.RTPCodec{
		audioCodec,
		videoCodec,
	}, bitrate)
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

	// if bitrate is 0, and estimator is enabled, use estimator bitrate
	if bitrate == 0 && estimator != nil {
		bitrate = estimator.GetTargetBitrate()
	}

	// audio track
	audioTrack, err := NewTrack(logger, audioCodec, connection)
	if err != nil {
		return nil, err
	}

	// set stream for audio track
	_, err = audioTrack.SetStream(audio)
	if err != nil {
		return nil, err
	}

	// if estimator is disabled, or in passive mode, disable video auto bitrate
	if !manager.config.EstimatorEnabled || manager.config.EstimatorPassive {
		videoAuto = false
	}

	videoRtcp := make(chan []rtcp.Packet, 1)

	// video track
	videoTrack, err := NewTrack(logger, videoCodec, connection,
		WithVideoAuto(videoAuto),
		WithRtcpChan(videoRtcp),
	)
	if err != nil {
		return nil, err
	}

	// let video stream bucket manager handle stream subscriptions
	video.SetReceiver(videoTrack)

	// data channel

	dataChannel, err := connection.CreateDataChannel("data", nil)
	if err != nil {
		return nil, err
	}

	peer := &WebRTCPeerCtx{
		logger:     logger,
		session:    session,
		metrics:    metrics,
		connection: connection,
		estimator:  estimator,
		// tracks & channels
		audioTrack:  audioTrack,
		videoTrack:  videoTrack,
		dataChannel: dataChannel,
		rtcpChannel: videoRtcp,
		// config
		iceTrickle:       manager.config.ICETrickle,
		estimatorPassive: manager.config.EstimatorPassive,
	}

	logger.Info().
		Int("target_bitrate", bitrate).
		Msg("estimated initial peer bitrate")

	// set initial video bitrate
	if err := peer.SetVideoBitrate(bitrate); err != nil {
		return nil, err
	}

	connection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		logger := logger.With().
			Str("kind", track.Kind().String()).
			Str("mime", track.Codec().RTPCodecCapability.MimeType).
			Logger()

		logger.Info().Msgf("received new remote track")

		if !session.Profile().CanShareMedia {
			err := receiver.Stop()
			logger.Warn().Err(err).Msg("media sharing is disabled for this session")
			return
		}

		// parse codec from remote track
		codec, ok := codec.ParseRTC(track.Codec())
		if !ok {
			err := receiver.Stop()
			logger.Warn().Err(err).Msg("remote track with unknown codec")
			return
		}

		var srcManager types.StreamSrcManager

		stopped := false
		stopFn := func() {
			if stopped {
				return
			}

			stopped = true
			err := receiver.Stop()
			srcManager.Stop()
			logger.Err(err).Msg("remote track stopped")
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
			err := receiver.Stop()
			logger.Warn().Err(err).Msg("remote track with unsupported codec type")
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
				err := connection.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{
						MediaSSRC: uint32(track.SSRC()),
					},
				})

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

	var once sync.Once
	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			session.SetWebRTCConnected(peer, true)
		case webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateFailed:
			connection.Close()
		case webrtc.PeerConnectionStateClosed:
			// ensure we only run this once
			once.Do(func() {
				session.SetWebRTCConnected(peer, false)
				if err = video.RemoveReceiver(videoTrack); err != nil {
					logger.Err(err).Msg("failed to remove video receiver")
				}
				audioTrack.Shutdown()
				videoTrack.Shutdown()
				close(videoRtcp)
			})
		}

		metrics.SetState(state)
	})

	dataChannel.OnOpen(func() {
		manager.curImage.AddListener(peer)
		manager.curPosition.AddListener(peer)

		// send initial cursor image
		cur, img, err := manager.curImage.GetCurrent()
		if err == nil {
			err := peer.SendCursorImage(cur, img)
			if err != nil {
				logger.Err(err).Msg("failed to set cursor image")
			}
		} else {
			logger.Err(err).Msg("failed to get cursor image")
		}

		// send initial cursor position
		x, y := manager.desktop.GetCursorPosition()
		err = peer.SendCursorPosition(x, y)
		if err != nil {
			logger.Err(err).Msg("failed to set cursor position")
		}
	})

	dataChannel.OnClose(func() {
		manager.curImage.RemoveListener(peer)
		manager.curPosition.RemoveListener(peer)
	})

	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
		if err := manager.handle(logger, message.Data, dataChannel, session); err != nil {
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

	// start metrics collectors
	go metrics.rtcpReceiver(videoRtcp)
	go metrics.connectionStats(connection)

	// start estimator reader
	go peer.estimatorReader()

	return offer, nil
}

func (manager *WebRTCManagerCtx) SetCursorPosition(x, y int) {
	manager.curPosition.Set(x, y)
}
