package webrtc

import (
	"errors"
	"io"
	"sync"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/types"
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
	}

	peer.listener = func(sample types.Sample) {
		if peer.paused {
			return
		}

		err := track.WriteSample(media.Sample(sample))
		if err != nil && errors.Is(err, io.ErrClosedPipe) {
			logger.Warn().Err(err).Msg("pipeline failed to write")
		}
	}

	err = peer.SetStream(stream)
	return peer, err
}

type PeerStreamTrack struct {
	logger   zerolog.Logger
	track    *webrtc.TrackLocalStaticSample
	paused   bool
	listener func(sample types.Sample)

	stream   types.StreamSinkManager
	streamMu sync.Mutex

	onRtcp   func(rtcp.Packet)
	onRtcpMu sync.RWMutex
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
			n, _, err := sender.Read(rtcpBuf)
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					return
				}

				peer.logger.Err(err).Msg("RTCP read error")
				continue
			}

			packets, err := rtcp.Unmarshal(rtcpBuf[:n])
			if err != nil {
				peer.logger.Err(err).Msg("RTCP unmarshal error")
				continue
			}

			peer.onRtcpMu.RLock()
			handler := peer.onRtcp
			peer.onRtcpMu.RUnlock()

			for _, packet := range packets {
				if handler != nil {
					go handler(packet)
				}
			}
		}
	}()

	return nil
}

func (peer *PeerStreamTrack) SetPaused(paused bool) {
	peer.paused = paused
}

func (peer *PeerStreamTrack) OnRTCP(f func(rtcp.Packet)) {
	peer.onRtcpMu.Lock()
	defer peer.onRtcpMu.Unlock()

	peer.onRtcp = f
}
