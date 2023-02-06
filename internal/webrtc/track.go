package webrtc

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
)

type Track struct {
	logger      zerolog.Logger
	track       *webrtc.TrackLocalStaticSample
	paused      bool
	videoAuto   bool
	videoAutoMu sync.RWMutex
	listener    func(sample types.Sample)

	stream   types.StreamSinkManager
	streamMu sync.Mutex

	onRtcp   func(rtcp.Packet)
	onRtcpMu sync.RWMutex

	bitrateChange func(int) (bool, error)
	videoChange   func(string) (bool, error)
}

type option func(*Track)

func WithVideoAuto(auto bool) option {
	return func(t *Track) {
		t.videoAuto = auto
	}
}

func NewTrack(logger zerolog.Logger, codec codec.RTPCodec, connection *webrtc.PeerConnection, opts ...option) (*Track, error) {
	id := codec.Type.String()
	track, err := webrtc.NewTrackLocalStaticSample(codec.Capability, id, "stream")
	if err != nil {
		return nil, err
	}

	logger = logger.With().Str("id", id).Logger()

	t := &Track{
		logger: logger,
		track:  track,
	}

	for _, opt := range opts {
		opt(t)
	}

	t.listener = func(sample types.Sample) {
		if t.paused {
			return
		}

		err := track.WriteSample(media.Sample(sample))
		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			logger.Warn().Err(err).Msg("failed to write sample to track")
		}
	}

	sender, err := connection.AddTrack(t.track)
	if err != nil {
		return nil, err
	}

	go t.rtcpReader(sender)

	return t, nil
}

func (t *Track) rtcpReader(sender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	for {
		n, _, err := sender.Read(rtcpBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				return
			}

			t.logger.Err(err).Msg("RTCP read error")
			continue
		}

		packets, err := rtcp.Unmarshal(rtcpBuf[:n])
		if err != nil {
			t.logger.Err(err).Msg("RTCP unmarshal error")
			continue
		}

		t.onRtcpMu.RLock()
		handler := t.onRtcp
		t.onRtcpMu.RUnlock()

		for _, packet := range packets {
			if handler != nil {
				go handler(packet)
			}
		}
	}
}

func (t *Track) SetStream(stream types.StreamSinkManager) (bool, error) {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	// if we already listen to the stream, do nothing
	if t.stream == stream {
		return false, nil
	}

	var err error
	if t.stream != nil {
		err = t.stream.MoveListenerTo(&t.listener, stream)
	} else {
		err = stream.AddListener(&t.listener)
	}
	if err != nil {
		return false, err
	}

	t.stream = stream

	return true, nil
}

func (t *Track) RemoveStream() {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	if t.stream != nil {
		_ = t.stream.RemoveListener(&t.listener)
		t.stream = nil
	}
}

func (t *Track) SetPaused(paused bool) {
	t.paused = paused
}

func (t *Track) OnRTCP(f func(rtcp.Packet)) {
	t.onRtcpMu.Lock()
	defer t.onRtcpMu.Unlock()

	t.onRtcp = f
}

func (t *Track) SetBitrate(bitrate int) (bool, error) {
	if t.bitrateChange == nil {
		return false, fmt.Errorf("bitrate change not supported")
	}

	return t.bitrateChange(bitrate)
}

func (t *Track) SetVideoID(videoID string) (bool, error) {
	if t.videoChange == nil {
		return false, fmt.Errorf("video change not supported")
	}

	return t.videoChange(videoID)
}

func (t *Track) OnBitrateChange(f func(bitrate int) (bool, error)) {
	t.bitrateChange = f
}

func (t *Track) OnVideoChange(f func(string) (bool, error)) {
	t.videoChange = f
}

func (t *Track) SetVideoAuto(auto bool) {
	t.videoAutoMu.Lock()
	defer t.videoAutoMu.Unlock()
	t.videoAuto = auto
}

func (t *Track) VideoAuto() bool {
	t.videoAutoMu.RLock()
	defer t.videoAutoMu.RUnlock()
	return t.videoAuto
}
