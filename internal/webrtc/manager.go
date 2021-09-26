package webrtc

import (
	"errors"
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

// the duration without network activity before a Agent is considered disconnected. Default is 5 Seconds
const disconnectedTimeout = 4 * time.Second

// the duration without network activity before a Agent is considered failed after disconnected. Default is 25 Seconds
const failedTimeout = 6 * time.Second

// how often the ICE Agent sends extra traffic if there is no activity, if media is flowing no traffic will be sent. Default is 2 seconds
const keepAliveInterval = 2 * time.Second

func New(desktop types.DesktopManager, capture types.CaptureManager, config *config.WebRTC) *WebRTCManagerCtx {
	return &WebRTCManagerCtx{
		logger: log.With().Str("module", "webrtc").Logger(),
		config: config,

		desktop:     desktop,
		capture:     capture,
		curImage:    cursor.NewImage(desktop),
		curPosition: cursor.NewPosition(desktop),

		participants: 0,
	}
}

type WebRTCManagerCtx struct {
	mu     sync.Mutex
	logger zerolog.Logger
	config *config.WebRTC

	desktop     types.DesktopManager
	capture     types.CaptureManager
	curImage    *cursor.ImageCtx
	curPosition *cursor.PositionCtx

	audioTrack    *webrtc.TrackLocalStaticSample
	audioListener func(sample types.Sample)
	participants  uint32
}

func (manager *WebRTCManagerCtx) Start() {
	var err error

	// create audio track
	audio := manager.capture.Audio()
	manager.audioTrack, err = webrtc.NewTrackLocalStaticSample(audio.Codec().Capability, "audio", "stream")
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio track")
	}

	manager.audioListener = func(sample types.Sample) {
		if err := manager.audioTrack.WriteSample(media.Sample(sample)); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				// The peerConnection has been closed.
				return
			}
			manager.logger.Warn().Err(err).Msg("audio pipeline failed to write")
		}
	}
	audio.AddListener(&manager.audioListener)

	manager.curImage.Start()

	manager.logger.Info().
		Bool("icelite", manager.config.ICELite).
		Bool("icetrickle", manager.config.ICETrickle).
		Interface("iceservers", manager.config.ICEServers).
		Str("nat1to1", strings.Join(manager.config.NAT1To1IPs, ",")).
		Str("epr", fmt.Sprintf("%d-%d", manager.config.EphemeralMin, manager.config.EphemeralMax)).
		Msg("webrtc starting")
}

func (manager *WebRTCManagerCtx) Shutdown() error {
	manager.logger.Info().Msg("shutdown")

	manager.curImage.Shutdown()
	manager.curPosition.Shutdown()

	audio := manager.capture.Audio()
	audio.RemoveListener(&manager.audioListener)

	return nil
}

func (manager *WebRTCManagerCtx) ICEServers() []types.ICEServer {
	return manager.config.ICEServers
}

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session, videoID string) (*webrtc.SessionDescription, error) {
	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()
	logger.Info().Msg("creating webrtc peer")

	// all videos must have the same codec
	videoStream, ok := manager.capture.Video(videoID)
	if !ok {
		return nil, types.ErrWebRTCVideoNotFound
	}

	connection, err := manager.newPeerConnection(videoStream.Codec(), logger)
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

	// create video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(videoStream.Codec().Capability, "video", "stream")
	if err != nil {
		return nil, err
	}

	videoListener := func(sample types.Sample) {
		if err := videoTrack.WriteSample(media.Sample(sample)); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				// The peerConnection has been closed.
				return
			}
			logger.Warn().Err(err).Msg("video pipeline failed to write")
		}
	}

	manager.mu.Lock()

	// should be stream started
	if videoStream.ListenersCount() == 0 {
		if err := videoStream.Start(); err != nil {
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
			return types.ErrWebRTCVideoNotFound
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

	peer := &WebRTCPeerCtx{
		logger:      logger,
		connection:  connection,
		dataChannel: dataChannel,
		changeVideo: changeVideo,
		iceTrickle:  manager.config.ICETrickle,
	}

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
			logger.Err(err).Msg("failed to get cursor image")
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

		if err = manager.handle(message.Data, session); err != nil {
			logger.Err(err).Msg("data handle failed")
		}
	})

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
	return peer.CreateOffer(false)
}
