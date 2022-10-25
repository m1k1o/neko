package webrtc

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/internal/webrtc/payload"
	"github.com/demodesk/neko/pkg/types"
)

type WebRTCPeerCtx struct {
	mu          sync.Mutex
	logger      zerolog.Logger
	connection  *webrtc.PeerConnection
	dataChannel *webrtc.DataChannel
	changeVideo func(bitrate int) error
	videoId     func() string
	setPaused   func(isPaused bool)
	iceTrickle  bool
}

func (peer *WebRTCPeerCtx) CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return nil, types.ErrWebRTCConnectionNotFound
	}

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

	if peer.connection == nil {
		return nil, types.ErrWebRTCConnectionNotFound
	}

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

func (peer *WebRTCPeerCtx) SetOffer(sdp string) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return types.ErrWebRTCConnectionNotFound
	}

	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	})
}

func (peer *WebRTCPeerCtx) SetAnswer(sdp string) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return types.ErrWebRTCConnectionNotFound
	}

	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeAnswer,
	})
}

func (peer *WebRTCPeerCtx) SetCandidate(candidate webrtc.ICECandidateInit) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return types.ErrWebRTCConnectionNotFound
	}

	return peer.connection.AddICECandidate(candidate)
}

func (peer *WebRTCPeerCtx) SetVideoBitrate(bitrate int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return types.ErrWebRTCConnectionNotFound
	}

	peer.logger.Info().Int("bitrate", bitrate).Msg("change video bitrate")
	return peer.changeVideo(bitrate)
}

// TODO: Refactor.
func (peer *WebRTCPeerCtx) GetVideoId() string {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.videoId()
}

func (peer *WebRTCPeerCtx) SetPaused(isPaused bool) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.connection == nil {
		return types.ErrWebRTCConnectionNotFound
	}

	peer.logger.Info().Bool("is_paused", isPaused).Msg("set paused")
	peer.setPaused(isPaused)
	return nil
}

func (peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.dataChannel == nil {
		return types.ErrWebRTCDataChannelNotFound
	}

	data := payload.CursorPosition{
		Header: payload.Header{
			Event:  payload.OP_CURSOR_POSITION,
			Length: 7,
		},
		X: uint16(x),
		Y: uint16(y),
	}

	buffer := &bytes.Buffer{}
	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}

func (peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage, img []byte) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.dataChannel == nil {
		return types.ErrWebRTCDataChannelNotFound
	}

	data := payload.CursorImage{
		Header: payload.Header{
			Event:  payload.OP_CURSOR_IMAGE,
			Length: uint16(11 + len(img)),
		},
		Width:  cur.Width,
		Height: cur.Height,
		Xhot:   cur.Xhot,
		Yhot:   cur.Yhot,
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, img); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
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
