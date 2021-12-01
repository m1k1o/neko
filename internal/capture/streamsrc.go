package capture

import (
	"errors"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/capture/gst"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
)

type StreamSrcManagerCtx struct {
	logger        zerolog.Logger
	codecPipeline map[string]string // codec -> pipeline

	codec       codec.RTPCodec
	pipeline    *gst.Pipeline
	pipelineMu  sync.Mutex
	pipelineStr string
}

func streamSrcNew(codecPipeline map[string]string, video_id string) *StreamSrcManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-src").
		Str("video_id", video_id).Logger()

	return &StreamSrcManagerCtx{
		logger:        logger,
		codecPipeline: codecPipeline,
	}
}

func (manager *StreamSrcManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.Stop()
}

func (manager *StreamSrcManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamSrcManagerCtx) Start(codec codec.RTPCodec) error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	found := false
	for codecName, pipeline := range manager.codecPipeline {
		if codecName == codec.Name {
			manager.pipelineStr = pipeline
			manager.codec = codec
			found = true
			break
		}
	}

	if !found {
		return errors.New("no pipeline found for a codec")
	}

	var err error

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", manager.pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(manager.pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.Play()
	return nil
}

func (manager *StreamSrcManagerCtx) Stop() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}

func (manager *StreamSrcManagerCtx) Push(bytes []byte) {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Push("src", bytes)
}

func (manager *StreamSrcManagerCtx) Started() bool {
	return manager.pipeline != nil
}
