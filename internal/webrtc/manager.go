package webrtc

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/webrtc/cursor"
)

// how long is can take between sending offer and connecting
const offerTimeout = 4 * time.Second

// the duration without network activity before a Agent is considered disconnected. Default is 5 Seconds
const disconnectedTimeout = 4 * time.Second

// the duration without network activity before a Agent is considered failed after disconnected. Default is 25 Seconds
const failedTimeout = 6 * time.Second

// how often the ICE Agent sends extra traffic if there is no activity, if media is flowing no traffic will be sent. Default is 2 seconds
const keepAliveInterval = 2 * time.Second

func New(desktop types.DesktopManager, capture types.CaptureManager, config *config.WebRTC) *WebRTCManagerCtx {
	return &WebRTCManagerCtx{
		logger:       log.With().Str("module", "webrtc").Logger(),
		desktop:      desktop,
		capture:      capture,
		config:       config,
		participants: 0,
		curImage:     cursor.NewImage(desktop),
		curPosition:  cursor.NewPosition(desktop),
	}
}

type WebRTCManagerCtx struct {
	mu           sync.Mutex
	logger       zerolog.Logger
	audioTrack   *webrtc.TrackLocalStaticSample
	audioStop    func()
	desktop      types.DesktopManager
	capture      types.CaptureManager
	config       *config.WebRTC
	participants uint32
	curImage     *cursor.ImageCtx
	curPosition  *cursor.PositionCtx
}

func (manager *WebRTCManagerCtx) Start() {
	var err error

	// create audio track
	audio := manager.capture.Audio()
	manager.audioTrack, err = webrtc.NewTrackLocalStaticSample(audio.Codec().Capability, "audio", "stream")
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio track")
	}

	audioListener := func(sample types.Sample) {
		if err := manager.audioTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("audio pipeline failed to write")
		}
	}

	audio.AddListener(&audioListener)
	manager.audioStop = func() {
		audio.RemoveListener(&audioListener)
	}

	manager.logger.Info().
		Str("ice_lite", fmt.Sprintf("%t", manager.config.ICELite)).
		Str("ice_trickle", fmt.Sprintf("%t", manager.config.ICETrickle)).
		Str("ice_servers", strings.Join(manager.config.ICEServers, ",")).
		Str("ephemeral_port_range", fmt.Sprintf("%d-%d", manager.config.EphemeralMin, manager.config.EphemeralMax)).
		Str("nat_ips", strings.Join(manager.config.NAT1To1IPs, ",")).
		Msgf("webrtc starting")

	manager.curImage.Start()
}

func (manager *WebRTCManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("webrtc shutting down")

	manager.curImage.Shutdown()
	manager.curPosition.Shutdown()

	manager.audioStop()
	return nil
}

func (manager *WebRTCManagerCtx) ICELite() bool {
	return manager.config.ICELite
}

func (manager *WebRTCManagerCtx) ICEServers() []string {
	return manager.config.ICEServers
}

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session, videoID string) (*webrtc.SessionDescription, error) {
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()

	// Create MediaEngine
	engine, err := manager.mediaEngine(videoID)
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
	if manager.config.ICETrickle {
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
	}

	// create video track
	videoStream, ok := manager.capture.Video(videoID)
	if !ok {
		manager.logger.Warn().Str("videoID", videoID).Msg("video stream not found")
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(videoStream.Codec().Capability, "video", "stream")
	if err != nil {
		manager.logger.Warn().Err(err).Msg("unable to create video track")
		return nil, err
	}

	videoListener := func(sample types.Sample) {
		if err := videoTrack.WriteSample(media.Sample(sample)); err != nil && err != io.ErrClosedPipe {
			manager.logger.Warn().Err(err).Msg("video pipeline failed to write")
		}
	}

	manager.mu.Lock()

	// should be stream started
	if videoStream.ListenersCount() == 0 {
		if err := videoStream.Start(); err != nil {
			manager.logger.Warn().Err(err).Msg("unable to start video pipeline")
			return nil, err
		}
	}

	videoStream.AddListener(&videoListener)

	// start audio, when first participant connects
	if !manager.capture.Audio().Started() {
		if err := manager.capture.Audio().Start(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to start audio stream")
		}
	}

	manager.participants = manager.participants + 1
	manager.mu.Unlock()

	changeVideo := func(videoID string) error {
		newVideoStream, ok := manager.capture.Video(videoID)
		if !ok {
			return fmt.Errorf("video stream not found")
		}

		// should be new stream started
		if newVideoStream.ListenersCount() == 0 {
			if err := newVideoStream.Start(); err != nil {
				return err
			}
		}

		// switch videoListeners
		videoStream.RemoveListener(&videoListener)
		newVideoStream.AddListener(&videoListener)

		// should be old stream stopped
		if videoStream.ListenersCount() == 0 {
			videoStream.Stop()
		}

		videoStream = newVideoStream
		return nil
	}

	rtpAudio, err := connection.AddTrack(manager.audioTrack)
	if err != nil {
		return nil, err
	}

	rtpVideo, err := connection.AddTrack(videoTrack)
	if err != nil {
		return nil, err
	}

	dataChannel, err := connection.CreateDataChannel("data", nil)
	if err != nil {
		return nil, err
	}

	offer, err := connection.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	if !manager.config.ICETrickle {
		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(connection)

		if err := connection.SetLocalDescription(offer); err != nil {
			return nil, err
		}

		<-gatherComplete
	} else {
		if err := connection.SetLocalDescription(offer); err != nil {
			return nil, err
		}
	}

	peer := &WebRTCPeerCtx{
		api:         api,
		connection:  connection,
		changeVideo: changeVideo,
		dataChannel: dataChannel,
	}

	cursorImage := func(entry *cursor.ImageEntry) {
		if err := peer.SendCursorImage(entry.Cursor, entry.Image); err != nil {
			manager.logger.Warn().Err(err).Msg("could not send cursor image")
		}
	}

	cursorPosition := func(x, y int) {
		if session.IsHost() {
			return
		}

		if err := peer.SendCursorPosition(x, y); err != nil {
			manager.logger.Warn().Err(err).Msg("could not send cursor position")
		}
	}

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			session.SetWebRTCConnected(peer, true)
		case webrtc.PeerConnectionStateDisconnected:
			fallthrough
		case webrtc.PeerConnectionStateFailed:
			connection.Close()
		case webrtc.PeerConnectionStateClosed:
			manager.mu.Lock()

			session.SetWebRTCConnected(peer, false)
			videoStream.RemoveListener(&videoListener)

			// should be stream stopped
			if videoStream.ListenersCount() == 0 {
				videoStream.Stop()
			}

			// decrease participants
			manager.participants = manager.participants - 1

			// stop audio, if last participant disonnects
			if manager.participants <= 0 {
				manager.participants = 0

				if manager.capture.Audio().Started() {
					manager.capture.Audio().Stop()
				}
			}

			manager.mu.Unlock()
		}
	})

	dataChannel.OnOpen(func() {
		manager.curImage.AddListener(&cursorImage)
		manager.curPosition.AddListener(&cursorPosition)

		// send initial cursor image
		entry, err := manager.curImage.Get()
		if err == nil {
			cursorImage(entry)
		} else {
			manager.logger.Warn().Err(err).Msg("failed to get cursor image")
		}

		// send initial cursor position
		x, y := manager.curPosition.Get()
		cursorPosition(x, y)
	})

	dataChannel.OnClose(func() {
		manager.curImage.RemoveListener(&cursorImage)
		manager.curPosition.RemoveListener(&cursorPosition)
	})

	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
		if !session.IsHost() {
			return
		}

		if err = manager.handle(message); err != nil {
			logger.Warn().Err(err).Msg("data handle failed")
		}
	})

	// offer timeout
	go func() {
		time.Sleep(offerTimeout)

		// already disconnected
		if connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			return
		}

		// not connected
		if connection.ConnectionState() != webrtc.PeerConnectionStateConnected {
			logger.Warn().Msg("connection timeouted, closing")
			connection.Close()
		}
	}()

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := rtpAudio.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := rtpVideo.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	session.SetWebRTCPeer(peer)
	return connection.LocalDescription(), nil
}

func (manager *WebRTCManagerCtx) mediaEngine(videoID string) (*webrtc.MediaEngine, error) {
	engine := &webrtc.MediaEngine{}

	// all videos must have the same codec
	video, ok := manager.capture.Video(videoID)
	if !ok {
		return nil, fmt.Errorf("default video track not found")
	}

	videoCodec := video.Codec()
	if err := videoCodec.Register(engine); err != nil {
		return nil, err
	}

	audioCodec := manager.capture.Audio().Codec()
	if err := audioCodec.Register(engine); err != nil {
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
	settings.SetICETimeouts(disconnectedTimeout, failedTimeout, keepAliveInterval)
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
