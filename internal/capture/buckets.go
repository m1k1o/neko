package capture

import (
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
)

type BucketsManagerCtx struct {
	logger    zerolog.Logger
	codec     codec.RTPCodec
	streams   map[string]*StreamSinkManagerCtx
	streamIDs []string
}

func bucketsNew(codec codec.RTPCodec, streams map[string]*StreamSinkManagerCtx, streamIDs []string) *BucketsManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "buckets").
		Logger()

	return &BucketsManagerCtx{
		logger:    logger,
		codec:     codec,
		streams:   streams,
		streamIDs: streamIDs,
	}
}

func (m *BucketsManagerCtx) shutdown() {
	m.logger.Info().Msgf("shutdown")
}

func (m *BucketsManagerCtx) destroyAll() {
	for _, stream := range m.streams {
		if stream.Started() {
			stream.destroyPipeline()
		}
	}
}

func (m *BucketsManagerCtx) recreateAll() error {
	for _, stream := range m.streams {
		if stream.Started() {
			err := stream.createPipeline()
			if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
				return err
			}
		}
	}

	return nil
}

func (m *BucketsManagerCtx) IDs() []string {
	return m.streamIDs
}

func (m *BucketsManagerCtx) Codec() codec.RTPCodec {
	return m.codec
}

func (m *BucketsManagerCtx) SetReceiver(receiver types.Receiver) error {
	receiver.OnBitrateChange(func(bitrate int) error {
		stream, ok := m.findNearestStream(bitrate)
		if !ok {
			return fmt.Errorf("no stream found for bitrate %d", bitrate)
		}

		return receiver.SetStream(stream)
	})

	return nil
}

func (m *BucketsManagerCtx) findNearestStream(bitrate int) (ss *StreamSinkManagerCtx, ok bool) {
	minDiff := math.MaxInt
	for _, s := range m.streams {
		streamBitrate, err := s.Bitrate()
		if err != nil {
			m.logger.Error().Err(err).Msgf("failed to get bitrate for stream %s", s.ID())
			continue
		}

		diffAbs := int(math.Abs(float64(bitrate - streamBitrate)))

		if diffAbs < minDiff {
			minDiff, ss = diffAbs, s
		}
	}
	ok = ss != nil
	return
}

func (m *BucketsManagerCtx) RemoveReceiver(receiver types.Receiver) error {
	receiver.OnBitrateChange(nil)
	receiver.RemoveStream()
	return nil
}
