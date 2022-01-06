package webrtc

import (
	"fmt"
	"strings"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
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

// send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
const rtcpPLIInterval = 3 * time.Second

func New(desktop types.DesktopManager, capture types.CaptureManager, config *config.WebRTC) *WebRTCManagerCtx {
	return &WebRTCManagerCtx{
		logger: log.With().Str("module", "webrtc").Logger(),
		config: config,

		desktop:     desktop,
		capture:     capture,
		curImage:    cursor.NewImage(desktop),
		curPosition: cursor.NewPosition(),
	}
}

type WebRTCManagerCtx struct {
	logger zerolog.Logger
	config *config.WebRTC

	desktop     types.DesktopManager
	capture     types.CaptureManager
	curImage    *cursor.ImageCtx
	curPosition *cursor.PositionCtx
}

func (manager *WebRTCManagerCtx) Start() {
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

	return nil
}

func (manager *WebRTCManagerCtx) ICEServers() []types.ICEServer {
	return manager.config.ICEServers
}

func (manager *WebRTCManagerCtx) CreatePeer(session types.Session, videoID string) (types.WebRTCPeer, error) {
	// add session id to logger context
	logger := manager.logger.With().Str("session_id", session.ID()).Logger()
	logger.Info().Msg("creating webrtc peer")

	// all audios must have the same codec
	audioStream := manager.capture.Audio()

	// all videos must have the same codec
	videoStream, ok := manager.capture.Video(videoID)
	if !ok {
		return nil, types.ErrWebRTCVideoNotFound
	}

	connection, err := manager.newPeerConnection([]codec.RTPCodec{
		audioStream.Codec(),
		videoStream.Codec(),
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

	audioTrack, err := manager.newPeerStreamTrack(audioStream, logger)
	if err != nil {
		return nil, err
	}

	if err := audioTrack.AddToConnection(connection); err != nil {
		return nil, err
	}

	// video track

	videoTrack, err := manager.newPeerStreamTrack(videoStream, logger)
	if err != nil {
		return nil, err
	}

	if err := videoTrack.AddToConnection(connection); err != nil {
		return nil, err
	}

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
			videoStream, ok := manager.capture.Video(videoID)
			if !ok {
				return types.ErrWebRTCVideoNotFound
			}

			return videoTrack.SetStream(videoStream)
		},
		iceTrickle: manager.config.ICETrickle,
	}

	connection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		defer receiver.Stop()

		logger := logger.With().
			Str("kind", track.Kind().String()).
			Str("mime", track.Codec().RTPCodecCapability.MimeType).
			Logger()

		logger.Info().Msgf("received new remote track")

		if !session.Profile().CanShareMedia {
			logger.Warn().Msg("media sharing is disabled for this session")
			return
		}

		// parse codec from remote track
		codec, ok := codec.ParseRTC(track.Codec())
		if !ok {
			logger.Warn().Msg("remote track with unknown codec")
			return
		}

		var srcManager types.StreamSrcManager
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			// audio -> microphone
			srcManager = manager.capture.Microphone()
		} else if track.Kind() == webrtc.RTPCodecTypeVideo {
			// video -> webcam
			srcManager = manager.capture.Webcam()
		}

		err := srcManager.Start(codec)
		if err != nil {
			logger.Err(err).Msg("failed to start pipeline")
			return
		}
		defer srcManager.Stop()

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
		logger.Info().Interface("data-channel", dc).Msg("got remote data channel")
	})

	connection.OnNegotiationNeeded(func() {
		logger.Warn().Msg("negotiation is needed")

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

	connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			session.SetWebRTCConnected(peer, true)
		case webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateFailed:
			connection.Close()
		case webrtc.PeerConnectionStateClosed:
			session.SetWebRTCConnected(peer, false)
			videoTrack.RemoveStream()
			audioTrack.RemoveStream()
		}
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
	return peer, nil
}
