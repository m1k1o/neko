package capture

import (
	"errors"

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
	for _, video := range m.streams {
		if video.Started() {
			video.destroyPipeline()
		}
	}
}

func (m *BucketsManagerCtx) recreateAll() error {
	for _, video := range m.streams {
		if video.Started() {
			err := video.createPipeline()
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
	receiver.OnVideoIdChange(func(videoID string) error {
		videoStream, ok := m.streams[videoID]
		if !ok {
			return types.ErrWebRTCVideoNotFound
		}

		return receiver.SetStream(videoStream)
	})

	// TODO: Save receiver.
	return nil
}

func (m *BucketsManagerCtx) RemoveReceiver(receiver types.Receiver) error {
	// TODO: Unsubribe from OnVideoIdChange.
	// TODO: Remove receiver.
	receiver.RemoveStream()
	return nil
}
