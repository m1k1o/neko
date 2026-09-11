package capture

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

var testStreamID atomic.Uint64

type fakeSinkPipeline struct {
	samples     chan types.Sample
	onDestroy   func()
	keyframes   int
	destroyOnce sync.Once
}

func newFakeSinkPipeline(onDestroy func()) *fakeSinkPipeline {
	return &fakeSinkPipeline{samples: make(chan types.Sample), onDestroy: onDestroy}
}

func (pipeline *fakeSinkPipeline) Sample() chan types.Sample { return pipeline.samples }
func (pipeline *fakeSinkPipeline) AttachAppsink(string)      {}
func (pipeline *fakeSinkPipeline) Play()                     {}
func (pipeline *fakeSinkPipeline) EmitVideoKeyframe() bool {
	pipeline.keyframes++
	return true
}
func (pipeline *fakeSinkPipeline) Destroy() {
	pipeline.destroyOnce.Do(func() {
		if pipeline.onDestroy != nil {
			pipeline.onDestroy()
		}
		close(pipeline.samples)
	})
}

type sampleRecorder struct {
	mu      sync.Mutex
	samples []types.Sample
}

type sampleDiscarder struct{}

func (sampleDiscarder) WriteSample(types.Sample) {}

func (recorder *sampleRecorder) WriteSample(sample types.Sample) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.samples = append(recorder.samples, sample)
}

func (recorder *sampleRecorder) count() int {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return len(recorder.samples)
}

func newTestStream(t *testing.T, streamCodec codec.RTPCodec, factory func(string) (sinkPipeline, error)) *StreamSinkManagerCtx {
	t.Helper()

	id := fmt.Sprintf("test-%d", testStreamID.Add(1))
	stream := streamSinkNew(streamCodec, func() (string, error) { return id, nil }, id)
	stream.pipelineFactory = factory
	t.Cleanup(stream.shutdown)
	return stream
}

func successfulPipelineFactory(created, destroyed *int) func(string) (sinkPipeline, error) {
	return func(string) (sinkPipeline, error) {
		*created++
		return newFakeSinkPipeline(func() { *destroyed++ }), nil
	}
}

func TestSubscribeRejectsNilConsumer(t *testing.T) {
	stream := newTestStream(t, codec.Opus(), successfulPipelineFactory(new(int), new(int)))

	if _, err := stream.Subscribe(nil); err == nil {
		t.Fatal("Subscribe(nil) returned no error")
	}
	if stream.started() {
		t.Fatal("failed subscription started the stream")
	}
}

func TestSubscriptionLifecycle(t *testing.T) {
	tests := []struct {
		name                string
		subscriptions       int
		closeSubscriptions  int
		repeatLastClose     bool
		wantCreates         int
		wantDestroys        int
		wantStreamRemaining bool
	}{
		{name: "first subscription starts pipeline", subscriptions: 1, wantCreates: 1, wantStreamRemaining: true},
		{name: "multiple subscriptions share pipeline", subscriptions: 2, wantCreates: 1, wantStreamRemaining: true},
		{name: "closing one keeps pipeline active", subscriptions: 2, closeSubscriptions: 1, wantCreates: 1, wantStreamRemaining: true},
		{name: "closing final stops pipeline", subscriptions: 2, closeSubscriptions: 2, wantCreates: 1, wantDestroys: 1},
		{name: "repeated close is harmless", subscriptions: 1, closeSubscriptions: 1, repeatLastClose: true, wantCreates: 1, wantDestroys: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created, destroyed := 0, 0
			stream := newTestStream(t, codec.Opus(), successfulPipelineFactory(&created, &destroyed))
			subscriptions := make([]types.StreamSubscription, 0, test.subscriptions)

			for range test.subscriptions {
				subscription, err := stream.Subscribe(&sampleRecorder{})
				if err != nil {
					t.Fatalf("Subscribe() error = %v", err)
				}
				subscriptions = append(subscriptions, subscription)
			}
			for i := range test.closeSubscriptions {
				if err := subscriptions[i].Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			}
			if test.repeatLastClose {
				if err := subscriptions[test.closeSubscriptions-1].Close(); err != nil {
					t.Fatalf("repeated Close() error = %v", err)
				}
			}

			if created != test.wantCreates || destroyed != test.wantDestroys {
				t.Fatalf("pipeline lifecycle = (%d creates, %d destroys), want (%d, %d)", created, destroyed, test.wantCreates, test.wantDestroys)
			}
			if stream.started() != test.wantStreamRemaining {
				t.Fatalf("stream started = %v, want %v", stream.started(), test.wantStreamRemaining)
			}

			for _, subscription := range subscriptions {
				_ = subscription.Close()
			}
		})
	}
}

func TestVideoSubscriptionWaitsForKeyframe(t *testing.T) {
	stream := newTestStream(t, codec.VP8(), successfulPipelineFactory(new(int), new(int)))
	recorder := &sampleRecorder{}
	subscription, err := stream.Subscribe(recorder)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	defer subscription.Close()

	stream.onSample(types.Sample{DeltaUnit: true})
	if recorder.count() != 0 {
		t.Fatal("video consumer received a delta unit before a keyframe")
	}

	stream.onSample(types.Sample{DeltaUnit: false})
	stream.onSample(types.Sample{DeltaUnit: true})
	if recorder.count() != 2 {
		t.Fatalf("video consumer received %d samples after keyframe, want 2", recorder.count())
	}
}

func TestAudioSubscriptionDoesNotWaitForKeyframe(t *testing.T) {
	stream := newTestStream(t, codec.Opus(), successfulPipelineFactory(new(int), new(int)))
	recorder := &sampleRecorder{}
	subscription, err := stream.Subscribe(recorder)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	defer subscription.Close()

	stream.onSample(types.Sample{DeltaUnit: true})
	if recorder.count() != 1 {
		t.Fatalf("audio consumer received %d samples, want 1", recorder.count())
	}
}

func TestSampleDispatchDoesNotAllocate(t *testing.T) {
	stream := newTestStream(t, codec.Opus(), successfulPipelineFactory(new(int), new(int)))
	subscription, err := stream.Subscribe(sampleDiscarder{})
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	defer subscription.Close()

	sample := types.Sample{Timestamp: time.Unix(0, 0)}
	stream.onSample(sample)
	if allocations := testing.AllocsPerRun(100, func() { stream.onSample(sample) }); allocations != 0 {
		t.Fatalf("allocations per dispatch = %f", allocations)
	}
}

func TestSubscriptionSwitch(t *testing.T) {
	t.Run("starts target before stopping source and waits for keyframe", func(t *testing.T) {
		var eventsMu sync.Mutex
		events := []string{}
		record := func(event string) {
			eventsMu.Lock()
			defer eventsMu.Unlock()
			events = append(events, event)
		}
		factory := func(name string) func(string) (sinkPipeline, error) {
			return func(string) (sinkPipeline, error) {
				record("create-" + name)
				return newFakeSinkPipeline(func() { record("destroy-" + name) }), nil
			}
		}

		source := newTestStream(t, codec.VP8(), factory("source"))
		target := newTestStream(t, codec.VP8(), factory("target"))
		recorder := &sampleRecorder{}
		subscription, err := source.Subscribe(recorder)
		if err != nil {
			t.Fatalf("Subscribe() error = %v", err)
		}
		defer subscription.Close()

		eventsMu.Lock()
		events = nil
		eventsMu.Unlock()
		if err := subscription.Switch(target); err != nil {
			t.Fatalf("Switch() error = %v", err)
		}

		eventsMu.Lock()
		gotEvents := append([]string(nil), events...)
		eventsMu.Unlock()
		wantEvents := []string{"create-target", "destroy-source"}
		if !reflect.DeepEqual(gotEvents, wantEvents) {
			t.Fatalf("switch events = %v, want %v", gotEvents, wantEvents)
		}

		target.onSample(types.Sample{DeltaUnit: true})
		if recorder.count() != 0 {
			t.Fatal("switched video consumer received a delta unit before a target keyframe")
		}
		target.onSample(types.Sample{DeltaUnit: false})
		if recorder.count() != 1 {
			t.Fatalf("switched video consumer received %d samples, want 1", recorder.count())
		}
	})

	t.Run("failed target startup leaves source active", func(t *testing.T) {
		source := newTestStream(t, codec.VP8(), successfulPipelineFactory(new(int), new(int)))
		target := newTestStream(t, codec.VP8(), func(string) (sinkPipeline, error) {
			return nil, errors.New("target startup failed")
		})
		recorder := &sampleRecorder{}
		subscription, err := source.Subscribe(recorder)
		if err != nil {
			t.Fatalf("Subscribe() error = %v", err)
		}
		defer subscription.Close()

		if err := subscription.Switch(target); err == nil {
			t.Fatal("Switch() returned no error")
		}
		if subscription.Stream() != source {
			t.Fatal("failed switch changed the subscription stream")
		}
		if !source.started() || target.started() {
			t.Fatalf("started state after failed switch = (source %v, target %v)", source.started(), target.started())
		}

		source.onSample(types.Sample{DeltaUnit: false})
		if recorder.count() != 1 {
			t.Fatal("source stopped delivering after failed switch")
		}
	})
}

func TestPipelineRecreationPreservesSubscription(t *testing.T) {
	created, destroyed := 0, 0
	stream := newTestStream(t, codec.VP8(), successfulPipelineFactory(&created, &destroyed))
	selector := streamSelectorNew(codec.VP8(), map[string]*StreamSinkManagerCtx{stream.ID(): stream}, []string{stream.ID()})
	recorder := &sampleRecorder{}
	subscription, err := stream.Subscribe(recorder)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	defer subscription.Close()

	selector.destroyPipelines()
	if subscription.Stream() != stream || !stream.started() {
		t.Fatal("destroying pipelines removed the logical subscription")
	}
	if err := selector.recreatePipelines(); err != nil {
		t.Fatalf("recreatePipelines() error = %v", err)
	}
	if created != 2 || destroyed != 1 {
		t.Fatalf("pipeline lifecycle = (%d creates, %d destroys), want (2, 1)", created, destroyed)
	}

	stream.onSample(types.Sample{DeltaUnit: false})
	if recorder.count() != 1 {
		t.Fatal("subscription did not receive samples after pipeline recreation")
	}
}

func TestPipelineRecreationWithNewSubscription(t *testing.T) {
	created, destroyed := 0, 0
	stream := newTestStream(t, codec.VP8(), successfulPipelineFactory(&created, &destroyed))
	selector := streamSelectorNew(codec.VP8(), map[string]*StreamSinkManagerCtx{stream.ID(): stream}, []string{stream.ID()})

	selector.destroyPipelines()
	// A new viewer can start a pipeline between the resize hooks.
	recorder := &sampleRecorder{}
	subscription, err := stream.Subscribe(recorder)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	defer subscription.Close()

	if err := selector.recreatePipelines(); err != nil {
		t.Fatalf("recreatePipelines() error = %v", err)
	}
	if created != 1 || destroyed != 0 {
		t.Fatalf("pipeline lifecycle = (%d creates, %d destroys), want (1, 0)", created, destroyed)
	}
	stream.onSample(types.Sample{DeltaUnit: false})
	if recorder.count() != 1 {
		t.Fatal("subscription did not receive samples from the existing pipeline")
	}
}
