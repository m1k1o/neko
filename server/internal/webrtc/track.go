package webrtc

import (
	"errors"
	"io"
	"sync"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type Track struct {
	logger zerolog.Logger
	track  *webrtc.TrackLocalStaticSample

	rtcpCh chan []rtcp.Packet
	sample chan types.Sample
	done   chan struct{}

	paused       bool
	stream       types.EncodedStream
	subscription types.StreamSubscription
	streamMu     sync.Mutex
	shutdownOnce sync.Once
}

type trackOption func(*Track)

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
		sample: make(chan types.Sample, 2),
		done:   make(chan struct{}),
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
	t.shutdownOnce.Do(func() {
		t.RemoveStream()
		close(t.done)
	})
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

// --- sample  ---

func (t *Track) sampleReader() {
	for {
		select {
		case <-t.done:
			t.logger.Debug().Msg("track sample reader closed")
			return
		case sample := <-t.sample:
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
}

func (t *Track) WriteSample(sample types.Sample) {
	select {
	case <-t.done:
		return
	default:
	}

	select {
	case <-t.done:
	case t.sample <- sample:
	default:
		t.logger.Trace().Msg("dropping sample: track channel full")
	}
}

// --- stream ---

func (t *Track) SetStream(stream types.EncodedStream) (bool, error) {
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

	if t.subscription != nil {
		if err := t.subscription.Switch(stream); err != nil {
			return false, err
		}
	} else {
		subscription, err := stream.Subscribe(t)
		if err != nil {
			return false, err
		}
		t.subscription = subscription
	}

	t.stream = stream
	return true, nil
}

func (t *Track) RemoveStream() {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	if t.stream == nil {
		t.stream = nil
		return
	}

	if t.subscription != nil {
		if err := t.subscription.Close(); err != nil {
			t.logger.Warn().Err(err).Msg("failed to close stream subscription")
		}
		t.subscription = nil
	}

	t.stream = nil
}

func (t *Track) Stream() (types.EncodedStream, bool) {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	return t.stream, t.stream != nil
}

// --- paused ---

func (t *Track) SetPaused(paused bool) {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	// if there is no state change or no stream, do nothing
	if t.paused == paused || t.stream == nil {
		t.paused = paused
		return
	}

	if paused {
		if t.subscription != nil {
			if err := t.subscription.Close(); err != nil {
				t.logger.Warn().Err(err).Msg("failed to pause stream subscription")
				return
			}
			t.subscription = nil
		}
	} else {
		subscription, err := t.stream.Subscribe(t)
		if err != nil {
			t.logger.Warn().Err(err).Msg("failed to resume stream subscription")
			return
		}
		t.subscription = subscription
	}

	t.paused = paused
}

func (t *Track) Paused() bool {
	t.streamMu.Lock()
	defer t.streamMu.Unlock()

	return t.paused
}
