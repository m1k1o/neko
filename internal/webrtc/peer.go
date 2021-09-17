package webrtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"
)

type WebRTCPeerCtx struct {
	mu          sync.Mutex
	logger      zerolog.Logger
	connection  *webrtc.PeerConnection
	dataChannel *webrtc.DataChannel
	changeVideo func(videoID string) error
	iceTrickle  bool
}

func (peer *WebRTCPeerCtx) CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	offer, err := peer.connection.CreateOffer(&webrtc.OfferOptions{
		ICERestart: ICERestart,
	})
	if err != nil {
		return nil, err
	}

	if !peer.iceTrickle {
		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peer.connection)

		if err := peer.connection.SetLocalDescription(offer); err != nil {
			return nil, err
		}

		<-gatherComplete
	} else {
		if err := peer.connection.SetLocalDescription(offer); err != nil {
			return nil, err
		}
	}

	return peer.connection.LocalDescription(), nil
}

func (peer *WebRTCPeerCtx) SignalAnswer(sdp string) error {
	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeAnswer,
	})
}

func (peer *WebRTCPeerCtx) SignalCandidate(candidate webrtc.ICECandidateInit) error {
	return peer.connection.AddICECandidate(candidate)
}

func (peer *WebRTCPeerCtx) SetVideoID(videoID string) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	peer.logger.Info().Str("video_id", videoID).Msg("change video id")
	return peer.changeVideo(videoID)
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
