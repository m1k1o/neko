package capture

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

var moveSinkSubscriptionMu sync.Mutex

type sinkPipeline interface {
	Sample() chan types.Sample
	AttachAppsink(string)
	Play()
	Destroy()
	EmitVideoKeyframe() bool
}

type StreamSinkManagerCtx struct {
	id string

	// wait for a keyframe before sending samples
	waitForKf bool

	bitrate   uint64
	brBuckets map[int]float64

	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	codec           codec.RTPCodec
	pipeline        sinkPipeline
	pipelineMu      sync.Mutex
	pipelineFn      func() (string, error)
	pipelineFactory func(string) (sinkPipeline, error)

	listeners   map[*streamSubscription]types.SampleConsumer
	listenersKf map[*streamSubscription]types.SampleConsumer // keyframe lobby
	listenersMu sync.Mutex

	// metrics
	currentListeners prometheus.Gauge
	totalBytes       prometheus.Counter
	pipelinesCounter prometheus.Counter
	pipelinesActive  prometheus.Gauge
}

func streamSinkNew(codec codec.RTPCodec, pipelineFn func() (string, error), id string) *StreamSinkManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-sink").
		Str("id", id).Logger()

	manager := &StreamSinkManagerCtx{
		id: id,

		// only wait for keyframes if the codec is video
		waitForKf: codec.IsVideo(),

		bitrate:   0,
		brBuckets: map[int]float64{},

		logger:     logger,
		codec:      codec,
		pipelineFn: pipelineFn,
		pipelineFactory: func(src string) (sinkPipeline, error) {
			return gst.CreatePipeline(src)
		},

		listeners:   map[*streamSubscription]types.SampleConsumer{},
		listenersKf: map[*streamSubscription]types.SampleConsumer{},

		// metrics
		currentListeners: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "streamsink_listeners",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Current number of listeners for a pipeline.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		totalBytes: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "streamsink_bytes",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of bytes created by the pipeline.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		pipelinesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "pipelines_total",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of created pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "streamsink",
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		pipelinesActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "pipelines_active",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of active pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "streamsink",
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
	}

	return manager
}

func (manager *StreamSinkManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.listenersMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	for key := range manager.listenersKf {
		delete(manager.listenersKf, key)
	}
	manager.listenersMu.Unlock()

	manager.destroyPipeline()
	manager.wg.Wait()
}

func (manager *StreamSinkManagerCtx) ID() string {
	return manager.id
}

func (manager *StreamSinkManagerCtx) Bitrate() uint64 {
	return manager.bitrate
}

func (manager *StreamSinkManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

type streamSubscription struct {
	mu       sync.Mutex
	stream   *StreamSinkManagerCtx
	consumer types.SampleConsumer
	closed   bool
}

func (subscription *streamSubscription) Stream() types.EncodedStream {
	subscription.mu.Lock()
	defer subscription.mu.Unlock()

	if subscription.closed {
		return nil
	}
	return subscription.stream
}

func (subscription *streamSubscription) Switch(stream types.EncodedStream) error {
	subscription.mu.Lock()
	defer subscription.mu.Unlock()

	if subscription.closed {
		return errors.New("stream subscription is closed")
	}

	target, ok := stream.(*StreamSinkManagerCtx)
	if !ok || target == nil {
		return errors.New("target stream is incompatible with this subscription")
	}
	if target == subscription.stream {
		return nil
	}

	if err := subscription.stream.moveSubscriptionTo(subscription, target); err != nil {
		return err
	}
	subscription.stream = target
	return nil
}

func (subscription *streamSubscription) Close() error {
	subscription.mu.Lock()
	defer subscription.mu.Unlock()

	if subscription.closed {
		return nil
	}

	subscription.stream.removeSubscription(subscription)
	subscription.closed = true
	subscription.stream = nil
	return nil
}

func (manager *StreamSinkManagerCtx) Subscribe(consumer types.SampleConsumer) (types.StreamSubscription, error) {
	if consumer == nil {
		return nil, errors.New("sample consumer cannot be nil")
	}

	subscription := &streamSubscription{stream: manager, consumer: consumer}
	if err := manager.addSubscription(subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (manager *StreamSinkManagerCtx) start() error {
	if len(manager.listeners)+len(manager.listenersKf) == 0 {
		err := manager.createPipeline()
		if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
			return err
		}

		manager.logger.Info().Msgf("first listener, starting")
	}

	return nil
}

func (manager *StreamSinkManagerCtx) stop() {
	if len(manager.listeners)+len(manager.listenersKf) == 0 {
		manager.destroyPipeline()
		manager.logger.Info().Msgf("last listener, stopping")
	}
}

func (manager *StreamSinkManagerCtx) addSubscriptionToMaps(subscription *streamSubscription) {
	emitKeyframe := false

	manager.listenersMu.Lock()
	if manager.waitForKf {
		// if this is the first listener, we need to emit a keyframe
		emitKeyframe = len(manager.listenersKf) == 0
		// if we're waiting for a keyframe, add it to the keyframe lobby
		manager.listenersKf[subscription] = subscription.consumer
	} else {
		// otherwise, add it as a regular listener
		manager.listeners[subscription] = subscription.consumer
	}
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("subscription", subscription).Msg("adding subscription")
	manager.currentListeners.Set(float64(manager.subscriptionsCount()))

	// if we will be waiting for a keyframe, emit one now
	if manager.pipeline != nil && emitKeyframe {
		manager.pipeline.EmitVideoKeyframe()
	}
}

func (manager *StreamSinkManagerCtx) removeSubscriptionFromMaps(subscription *streamSubscription) {
	manager.listenersMu.Lock()
	delete(manager.listeners, subscription)
	delete(manager.listenersKf, subscription)
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("subscription", subscription).Msg("removing subscription")
	manager.currentListeners.Set(float64(manager.subscriptionsCount()))
}

func (manager *StreamSinkManagerCtx) addSubscription(subscription *streamSubscription) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// start if stopped
	if err := manager.start(); err != nil {
		return err
	}

	// add listener
	manager.addSubscriptionToMaps(subscription)

	return nil
}

func (manager *StreamSinkManagerCtx) removeSubscription(subscription *streamSubscription) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.removeSubscriptionFromMaps(subscription)

	// stop if started
	manager.stop()

}

// moveSubscriptionTo starts the target before moving delivery and stopping the source.
func (manager *StreamSinkManagerCtx) moveSubscriptionTo(subscription *streamSubscription, targetStream *StreamSinkManagerCtx) error {
	// we need to acquire both mutextes, from source stream and from target stream
	// in order to do that safely (without possibility of deadlock) we need third
	// global mutex, that ensures atomic locking

	// lock global mutex
	moveSinkSubscriptionMu.Lock()

	// lock source stream
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// lock target stream
	targetStream.mu.Lock()
	defer targetStream.mu.Unlock()

	// unlock global mutex
	moveSinkSubscriptionMu.Unlock()

	// start if stopped
	if err := targetStream.start(); err != nil {
		return err
	}

	// swap listeners
	manager.removeSubscriptionFromMaps(subscription)
	targetStream.addSubscriptionToMaps(subscription)

	// stop if started
	manager.stop()

	return nil
}

func (manager *StreamSinkManagerCtx) subscriptionsCount() int {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	return len(manager.listeners) + len(manager.listenersKf)
}

func (manager *StreamSinkManagerCtx) started() bool {
	return manager.subscriptionsCount() > 0
}

func (manager *StreamSinkManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	pipelineStr, err := manager.pipelineFn()
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = manager.pipelineFactory(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.AttachAppsink("appsink")
	manager.pipeline.Play()

	pipeline := manager.pipeline
	manager.wg.Go(func() {
		manager.logger.Debug().Msg("started emitting samples")

		for {
			sample, ok := <-pipeline.Sample()
			if !ok {
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			}

			manager.onSample(sample)
		}
	})

	manager.pipelinesCounter.Inc()
	manager.pipelinesActive.Set(1)

	return nil
}

func (manager *StreamSinkManagerCtx) saveSampleBitrate(timestamp time.Time, delta float64) {
	// get unix timestamp in seconds
	sec := timestamp.Unix()
	// last bucket is timestamp rounded to 3 seconds - 1 second
	last := int((sec - 1) % 3)
	// current bucket is timestamp rounded to 3 seconds
	curr := int(sec % 3)
	// next bucket is timestamp rounded to 3 seconds + 1 second
	next := int((sec + 1) % 3)

	if manager.brBuckets[next] != 0 {
		// update bitrate, TODO: atomic?
		manager.bitrate = uint64(manager.brBuckets[last])
		// empty next bucket
		manager.brBuckets[next] = 0
	}

	// add rate to current bucket
	manager.brBuckets[curr] += delta
}

func (manager *StreamSinkManagerCtx) onSample(sample types.Sample) {
	manager.listenersMu.Lock()

	// save to metrics
	length := float64(sample.Length)
	manager.totalBytes.Add(length)
	manager.saveSampleBitrate(sample.Timestamp, length)

	// if is not delta unit -> it can be decoded independently -> it is a keyframe
	if manager.waitForKf && !sample.DeltaUnit && len(manager.listenersKf) > 0 {
		// if current sample is a keyframe, move listeners from
		// keyframe lobby to actual listeners map and clear lobby
		for k, v := range manager.listenersKf {
			manager.listeners[k] = v
		}
		manager.listenersKf = make(map[*streamSubscription]types.SampleConsumer)
	}

	// copy listeners before releasing lock to avoid holding it during dispatch
	listeners := make([]types.SampleConsumer, 0, len(manager.listeners))
	for _, l := range manager.listeners {
		listeners = append(listeners, l)
	}
	manager.listenersMu.Unlock()

	for _, l := range listeners {
		l.WriteSample(sample)
	}
}

func (manager *StreamSinkManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil

	manager.pipelinesActive.Set(0)

	manager.brBuckets = make(map[int]float64)
	manager.bitrate = 0
}
