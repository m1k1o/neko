package webrtc

import (
	"demodesk/neko/internal/types"
	"errors"
	"io"
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"
)

func (manager *WebRTCManagerCtx) newPeerTrack(stream types.StreamManager, logger zerolog.Logger) (*PeerTrack, error) {
	codec := stream.Codec()

	id := codec.Type.String()
	track, err := webrtc.NewTrackLocalStaticSample(codec.Capability, id, "stream")
	if err != nil {
		return nil, err
	}

	logger = logger.With().Str("id", id).Logger()

	peer := &PeerTrack{
		logger: logger,
		track:  track,
		listener: func(sample types.Sample) {
			err := track.WriteSample(media.Sample(sample))
			if err != nil && errors.Is(err, io.ErrClosedPipe) {
				logger.Warn().Err(err).Msg("pipeline failed to write")
			}
		},
	}

	peer.SetStream(stream)
	return peer, nil

}

type PeerTrack struct {
	logger   zerolog.Logger
	track    *webrtc.TrackLocalStaticSample
	listener func(sample types.Sample)

	streamMu sync.Mutex
	stream   types.StreamManager
}

func (peer *PeerTrack) SetStream(stream types.StreamManager) error {
	peer.streamMu.Lock()
	defer peer.streamMu.Unlock()

	// prepare new listener
	dispatcher, err := stream.NewListener(&peer.listener)
	if err != nil {
		return err
	}

	// remove previous listener (in case it existed)
	if peer.stream != nil {
		peer.stream.RemoveListener(&peer.listener)
	}

	// add new listener
	close(dispatcher)

	peer.stream = stream
	return nil
}

func (peer *PeerTrack) RemoveStream() {
	peer.streamMu.Lock()
	defer peer.streamMu.Unlock()

	if peer.stream != nil {
		peer.stream.RemoveListener(&peer.listener)
	}
}

func (peer *PeerTrack) AddToConnection(connection *webrtc.PeerConnection) error {
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
