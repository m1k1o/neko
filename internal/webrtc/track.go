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
	logger zerolog.Logger
	track  *webrtc.TrackLocalStaticSample

	rtcpCh chan []rtcp.Packet
	sample chan types.Sample

	videoAuto   bool
	videoAutoMu sync.RWMutex

	paused   bool
	stream   types.StreamSinkManager
	streamMu sync.Mutex

	bitrateChange func(int) (bool, error)
	videoChange   func(string) (bool, error)
}

type trackOption func(*Track)

func WithVideoAuto(auto bool) trackOption {
	return func(t *Track) {
		t.videoAuto = auto
	}
}

func WithRtcpChan(rtcp chan []rtcp.Packet) trackOption {
	return func(t *Track) {
		t.rtcpCh = rtcp
	}
}

func NewTrack(logger zerolog.Logger, codec codec.RTPCodec, connection *webrtc.PeerConnection, opts ...trackOption) (*Track, error) {
	id := codec.Type.String()
	track, err := webrtc.NewTrackLocalStaticSample(codec.Capability, id, "stream")
	if err != nil {
		return nil, err
	}

	t := &Track{
		logger: logger.With().Str("id", id).Logger(),
		track:  track,
		rtcpCh: nil,
		sample: make(chan types.Sample),
	}

	for _, opt := range opts {
		opt(t)
	}

	sender, err := connection.AddTrack(t.track)
	if err != nil {
		return nil, err
	}

	go t.rtcpReader(sender)
	go t.sampleReader()

	return t, nil
}

func (t *Track) Shutdown() {
	t.RemoveStream()
	close(t.sample)
}

func (t *Track) rtcpReader(sender *webrtc.RTPSender) {
	for {
		packets, _, err := sender.ReadRTCP()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				t.logger.Debug().Msg("track rtcp reader closed")
				return
			}

			t.logger.Warn().Err(err).Msg("failed to read track rtcp")
			continue
		}

		if t.rtcpCh != nil {
			t.rtcpCh <- packets
		}
	}
}

func (t *Track) sampleReader() {
	for {
		sample, ok := <-t.sample
		if !ok {
			t.logger.Debug().Msg("track sample reader closed")
			return
		}

		err := t.track.WriteSample(media.Sample{
			Data:      sample.Data,
			Duration:  sample.Duration,
			Timestamp: sample.Timestamp,
		})

		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			t.logger.Warn().Err(err).Msg("failed to write sample to track")
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

	// if paused, we switch the stream but don't add the listener
	if t.paused {
		t.stream = stream
		return true, nil
	}

	var err error
	if t.stream != nil {
		err = t.stream.MoveListenerTo(t, stream)
	} else {
		err = stream.AddListener(t)
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

	// if there is no stream, or paused we don't need to remove the listener
	if t.stream == nil || t.paused {
		t.stream = nil
		return
	}

	err := t.stream.RemoveListener(t)
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to remove listener from stream")
	}

	t.stream = nil
}

func (t *Track) SetPaused(paused bool) {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	// if there is no state change or no stream, do nothing
	if t.paused == paused || t.stream == nil {
		return
	}

	var err error
	if paused {
		err = t.stream.RemoveListener(t)
	} else {
		err = t.stream.AddListener(t)
	}
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to change listener state")
		return
	}

	t.paused = paused
}

func (t *Track) WriteSample(sample types.Sample) {
	t.sample <- sample
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
