package webrtc

import (
	"errors"
	"io"
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"

	"gitlab.com/demodesk/neko/server/pkg/types"
)

func (manager *WebRTCManagerCtx) newPeerStreamTrack(stream types.StreamSinkManager, logger zerolog.Logger) (*PeerStreamTrack, error) {
	codec := stream.Codec()

	id := codec.Type.String()
	track, err := webrtc.NewTrackLocalStaticSample(codec.Capability, id, "stream")
	if err != nil {
		return nil, err
	}

	logger = logger.With().Str("id", id).Logger()

	peer := &PeerStreamTrack{
		logger: logger,
		track:  track,
		listener: func(sample types.Sample) {
			err := track.WriteSample(media.Sample(sample))
			if err != nil && errors.Is(err, io.ErrClosedPipe) {
				logger.Warn().Err(err).Msg("pipeline failed to write")
			}
		},
	}

	err = peer.SetStream(stream)
	return peer, err
}

type PeerStreamTrack struct {
	logger   zerolog.Logger
	track    *webrtc.TrackLocalStaticSample
	listener func(sample types.Sample)

	stream   types.StreamSinkManager
	streamMu sync.Mutex
}

func (peer *PeerStreamTrack) SetStream(stream types.StreamSinkManager) error {
	peer.streamMu.Lock()
	defer peer.streamMu.Unlock()

	var err error
	if peer.stream != nil {
		err = peer.stream.MoveListenerTo(&peer.listener, stream)
	} else {
		err = stream.AddListener(&peer.listener)
	}

	if err == nil {
		peer.stream = stream
	}

	return err
}

func (peer *PeerStreamTrack) RemoveStream() {
	peer.streamMu.Lock()
	defer peer.streamMu.Unlock()

	if peer.stream != nil {
		_ = peer.stream.RemoveListener(&peer.listener)
		peer.stream = nil
	}
}

func (peer *PeerStreamTrack) AddToConnection(connection *webrtc.PeerConnection) error {
	sender, err := connection.AddTrack(peer.track)
	if err != nil {
		return err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := sender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	return nil
}
