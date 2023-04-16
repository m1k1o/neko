package webrtc

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/internal/webrtc/payload"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

type WebRTCPeerCtx struct {
	mu         sync.Mutex
	logger     zerolog.Logger
	session    types.Session
	metrics    *metrics
	connection *webrtc.PeerConnection
	estimator  cc.BandwidthEstimator
	// tracks & channels
	audioTrack  *Track
	videoTrack  *Track
	dataChannel *webrtc.DataChannel
	rtcpChannel chan []rtcp.Packet
	// config
	iceTrickle       bool
	estimatorPassive bool
}

//
// connection
//

func (peer *WebRTCPeerCtx) CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	offer, err := peer.connection.CreateOffer(&webrtc.OfferOptions{
		ICERestart: ICERestart,
	})
	if err != nil {
		return nil, err
	}

	return peer.setLocalDescription(offer)
}

func (peer *WebRTCPeerCtx) CreateAnswer() (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	answer, err := peer.connection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	return peer.setLocalDescription(answer)
}

func (peer *WebRTCPeerCtx) setLocalDescription(description webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if !peer.iceTrickle {
		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peer.connection)

		if err := peer.connection.SetLocalDescription(description); err != nil {
			return nil, err
		}

		<-gatherComplete
	} else {
		if err := peer.connection.SetLocalDescription(description); err != nil {
			return nil, err
		}
	}

	return peer.connection.LocalDescription(), nil
}

func (peer *WebRTCPeerCtx) SetRemoteDescription(desc webrtc.SessionDescription) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.connection.SetRemoteDescription(desc)
}

func (peer *WebRTCPeerCtx) SetCandidate(candidate webrtc.ICECandidateInit) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.connection.AddICECandidate(candidate)
}

func (peer *WebRTCPeerCtx) Destroy() {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection != nil {
		err := peer.connection.Close()
		peer.logger.Err(err).Msg("peer connection destroyed")
		peer.connection = nil
	}
}

func (peer *WebRTCPeerCtx) estimatorReader() {
	// if estimator is disabled, do nothing
	if peer.estimator == nil {
		return
	}

	// use a ticker to get current client target bitrate
	ticker := time.NewTicker(bitrateCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		targetBitrate := peer.estimator.GetTargetBitrate()
		peer.metrics.SetReceiverEstimatedTargetBitrate(float64(targetBitrate))

		if peer.connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			break
		}

		if !peer.videoTrack.VideoAuto() {
			continue
		}

		if !peer.estimatorPassive {
			err := peer.SetVideoBitrate(targetBitrate)
			if err != nil {
				peer.logger.Warn().Err(err).Msg("failed to set video bitrate")
			}
		}
	}
}

//
// video
//

func (peer *WebRTCPeerCtx) SetVideoBitrate(peerBitrate int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// when switching from manual to auto bitrate estimation, in case the estimator is
	// idle (lastBitrate > maxBitrate), we want to go back to the previous estimated bitrate
	if peerBitrate == 0 && peer.estimator != nil && !peer.estimatorPassive {
		peerBitrate = peer.estimator.GetTargetBitrate()
		peer.logger.Debug().
			Int("peer_bitrate", peerBitrate).
			Msg("evaluated bitrate")
	}

	changed, err := peer.videoTrack.SetBitrate(peerBitrate)
	if err != nil {
		return err
	}

	if !changed {
		// TODO: return error?
		return nil
	}

	videoID := peer.videoTrack.stream.ID()
	bitrate := peer.videoTrack.stream.Bitrate()

	peer.metrics.SetVideoID(videoID)
	peer.logger.Debug().
		Int("peer_bitrate", peerBitrate).
		Int("video_bitrate", bitrate).
		Str("video_id", videoID).
		Msg("peer bitrate triggered video stream change")

	go peer.session.Send(
		event.SIGNAL_VIDEO,
		message.SignalVideo{
			Video:     videoID,
			Bitrate:   bitrate,
			VideoAuto: peer.videoTrack.VideoAuto(),
		})

	return nil
}

func (peer *WebRTCPeerCtx) SetVideoID(videoID string) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	changed, err := peer.videoTrack.SetVideoID(videoID)
	if err != nil {
		return err
	}

	if !changed {
		// TODO: return error?
		return nil
	}

	bitrate := peer.videoTrack.stream.Bitrate()

	peer.logger.Debug().
		Str("video_id", videoID).
		Int("video_bitrate", bitrate).
		Msg("peer video id triggered video stream change")

	go peer.session.Send(
		event.SIGNAL_VIDEO,
		message.SignalVideo{
			Video:     videoID,
			Bitrate:   bitrate,
			VideoAuto: peer.videoTrack.VideoAuto(),
		})

	return nil
}

func (peer *WebRTCPeerCtx) GetVideoID() string {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// TODO: Refactor.
	return peer.videoTrack.stream.ID()
}

func (peer *WebRTCPeerCtx) SetPaused(isPaused bool) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	peer.logger.Info().Bool("is_paused", isPaused).Msg("set paused")
	peer.videoTrack.SetPaused(isPaused)
	peer.audioTrack.SetPaused(isPaused)
	return nil
}

func (peer *WebRTCPeerCtx) SetVideoAuto(videoAuto bool) {
	// if estimator is enabled and is not passive, enable video auto bitrate
	if peer.estimator != nil && !peer.estimatorPassive {
		peer.videoTrack.SetVideoAuto(videoAuto)
	} else {
		peer.logger.Warn().Msg("estimator is disabled or in passive mode, cannot change video auto")
		peer.videoTrack.SetVideoAuto(false) // ensure video auto is disabled
	}
}

func (peer *WebRTCPeerCtx) VideoAuto() bool {
	return peer.videoTrack.VideoAuto()
}

//
// data channel
//

func (peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// do not send cursor position to host
	if peer.session.IsHost() {
		return nil
	}

	header := payload.Header{
		Event:  payload.OP_CURSOR_POSITION,
		Length: 7,
	}

	data := payload.CursorPosition{
		X: uint16(x),
		Y: uint16(y),
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}

func (peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage, img []byte) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	header := payload.Header{
		Event:  payload.OP_CURSOR_IMAGE,
		Length: uint16(11 + len(img)),
	}

	data := payload.CursorImage{
		Width:  cur.Width,
		Height: cur.Height,
		Xhot:   cur.Xhot,
		Yhot:   cur.Yhot,
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, img); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}
