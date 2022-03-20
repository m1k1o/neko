package capture

import (
	"errors"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/demodesk/neko/server/internal/capture/gst"
	"gitlab.com/demodesk/neko/server/internal/types"
	"gitlab.com/demodesk/neko/server/internal/types/codec"
)

type StreamSrcManagerCtx struct {
	logger        zerolog.Logger
	enabled       bool
	codecPipeline map[string]string // codec -> pipeline

	codec       codec.RTPCodec
	pipeline    *gst.Pipeline
	pipelineMu  sync.Mutex
	pipelineStr string
}

func streamSrcNew(enabled bool, codecPipeline map[string]string, video_id string) *StreamSrcManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-src").
		Str("video_id", video_id).Logger()

	return &StreamSrcManagerCtx{
		logger:        logger,
		enabled:       enabled,
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

	if !manager.enabled {
		return errors.New("stream-src not enabled")
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

	manager.pipeline.AttachAppsrc("appsrc")
	manager.pipeline.Play()
	return nil
}

func (manager *StreamSrcManagerCtx) Stop() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}

func (manager *StreamSrcManagerCtx) Push(bytes []byte) {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Push(bytes)
}

func (manager *StreamSrcManagerCtx) Started() bool {
	return manager.pipeline != nil
}
