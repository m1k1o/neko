package capture

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/gst"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
)

type StreamSrcManagerCtx struct {
	logger        zerolog.Logger
	enabled       bool
	codecPipeline map[string]string // codec -> pipeline

	codec       codec.RTPCodec
	pipeline    gst.Pipeline
	pipelineMu  sync.Mutex
	pipelineStr string

	// metrics
	pushedData       map[string]prometheus.Summary
	pipelinesCounter map[string]prometheus.Counter
	pipelinesActive  map[string]prometheus.Gauge
}

func streamSrcNew(enabled bool, codecPipeline map[string]string, video_id string) *StreamSrcManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-src").
		Str("video_id", video_id).Logger()

	pushedData := map[string]prometheus.Summary{}
	pipelinesCounter := map[string]prometheus.Counter{}
	pipelinesActive := map[string]prometheus.Gauge{}

	for codecName, pipeline := range codecPipeline {
		codec, ok := codec.ParseStr(codecName)
		if !ok {
			logger.Fatal().
				Str("codec", codecName).
				Str("pipeline", pipeline).
				Msg("unknown codec name")
		}

		pushedData[codecName] = promauto.NewSummary(prometheus.SummaryOpts{
			Name:      "streamsrc_data_bytes",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Data pushed to a pipeline (in bytes).",
			ConstLabels: map[string]string{
				"video_id":   video_id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		})
		pipelinesCounter[codecName] = promauto.NewCounter(prometheus.CounterOpts{
			Name:      "pipelines_total",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of created pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "streamsrc",
				"video_id":   video_id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		})
		pipelinesActive[codecName] = promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "pipelines_active",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of active pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "streamsrc",
				"video_id":   video_id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		})
	}

	return &StreamSrcManagerCtx{
		logger:        logger,
		enabled:       enabled,
		codecPipeline: codecPipeline,

		// metrics
		pushedData:       pushedData,
		pipelinesCounter: pipelinesCounter,
		pipelinesActive:  pipelinesActive,
	}
}

func (manager *StreamSrcManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.Stop()
}

func (manager *StreamSrcManagerCtx) Codec() codec.RTPCodec {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

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

	manager.pipelinesCounter[manager.codec.Name].Inc()
	manager.pipelinesActive[manager.codec.Name].Set(1)

	return nil
}

func (manager *StreamSrcManagerCtx) Stop() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.pipeline = nil

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", manager.pipelineStr).
		Msgf("destroying pipeline")

	manager.pipelinesActive[manager.codec.Name].Set(0)
}

func (manager *StreamSrcManagerCtx) Push(bytes []byte) {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Push(bytes)
	manager.pushedData[manager.codec.Name].Observe(float64(len(bytes)))
}

func (manager *StreamSrcManagerCtx) Started() bool {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	return manager.pipeline != nil
}
