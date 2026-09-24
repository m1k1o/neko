package capture

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

var testStreamID atomic.Uint64

type fakeSinkPipeline struct {
	samples   chan types.Sample
	onDestroy func()
}

func (p *fakeSinkPipeline) Sample() chan types.Sample { return p.samples }
func (p *fakeSinkPipeline) AttachAppsink(string)      {}
func (p *fakeSinkPipeline) Play()                     {}
func (p *fakeSinkPipeline) EmitVideoKeyframe() bool   { return true }
func (p *fakeSinkPipeline) Destroy() {
	p.onDestroy()
	close(p.samples)
}

type testSink struct {
	*StreamSinkManagerCtx
	created, destroyed int
}

func newTestSink(t *testing.T, c codec.RTPCodec) *testSink {
	t.Helper()
	id := fmt.Sprintf("test-%d", testStreamID.Add(1))
	s := &testSink{StreamSinkManagerCtx: streamSinkNew(c, func() (string, error) { return id, nil }, id)}
	s.pipelineFactory = func(string) (sinkPipeline, error) {
		s.created++
		return &fakeSinkPipeline{make(chan types.Sample), func() { s.destroyed++ }}, nil
	}
	t.Cleanup(s.shutdown)
	return s
}

func (s *testSink) check(t *testing.T, created, destroyed int, started bool) {
	t.Helper()
	if s.created != created || s.destroyed != destroyed || s.started() != started {
		t.Fatalf("lifecycle = (%d, %d, %v), want (%d, %d, %v)", s.created, s.destroyed, s.started(), created, destroyed, started)
	}
}

func checkError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func subscribe(t *testing.T, s types.EncodedStream, consumer types.SampleConsumer) types.StreamSubscription {
	t.Helper()
	sub, err := s.Subscribe(consumer)
	checkError(t, err)
	t.Cleanup(func() { checkError(t, sub.Close()) })
	return sub
}

type sampleCounter struct{ atomic.Int64 }

func (c *sampleCounter) WriteSample(types.Sample) { c.Add(1) }

func TestSubscriptionLifecycle(t *testing.T) {
	s := newTestSink(t, codec.Opus())
	if _, err := s.Subscribe(nil); err == nil {
		t.Fatal("Subscribe(nil) succeeded")
	}
	s.check(t, 0, 0, false)

	a := subscribe(t, s, &sampleCounter{})
	s.check(t, 1, 0, true)
	b := subscribe(t, s, &sampleCounter{})
	s.check(t, 1, 0, true)
	checkError(t, a.Close())
	s.check(t, 1, 0, true)
	checkError(t, b.Close())
	s.check(t, 1, 1, false)
	checkError(t, b.Close())
	s.check(t, 1, 1, false)
}

func TestSampleDispatch(t *testing.T) {
	for _, c := range []codec.RTPCodec{codec.Opus(), codec.VP8()} {
		t.Run(c.Name, func(t *testing.T) {
			s := newTestSink(t, c)
			counter := &sampleCounter{}
			subscribe(t, s, counter)
			for i, delta := range []bool{true, false, true} {
				s.onSample(types.Sample{DeltaUnit: delta})
				want := int64(i + 1)
				if c.IsVideo() {
					want--
				}
				if got := counter.Load(); got != want {
					t.Fatalf("sample %d: deliveries = %d, want %d", i, got, want)
				}
			}
			sample := types.Sample{Timestamp: time.Unix(0, 0), DeltaUnit: true}
			if allocs := testing.AllocsPerRun(100, func() { s.onSample(sample) }); allocs != 0 {
				t.Fatalf("allocations per dispatch = %f, want 0", allocs)
			}
		})
	}
}

func TestSubscriptionSwitch(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("startup-fails=%v", fail), func(t *testing.T) {
			source, target := newTestSink(t, codec.VP8()), newTestSink(t, codec.VP8())
			factory := target.pipelineFactory
			target.pipelineFactory = func(src string) (sinkPipeline, error) {
				if source.destroyed != 0 {
					t.Error("source stopped before target startup")
				}
				if fail {
					return nil, errors.New("target startup failed")
				}
				return factory(src)
			}
			counter := &sampleCounter{}
			sub := subscribe(t, source, counter)
			err := sub.Switch(target.StreamSinkManagerCtx)
			if fail {
				if err == nil || sub.Stream() != source.StreamSinkManagerCtx {
					t.Fatal("failed switch did not preserve the source subscription")
				}
				source.check(t, 1, 0, true)
				target.check(t, 0, 0, false)
				source.onSample(types.Sample{})
			} else {
				checkError(t, err)
				if sub.Stream() != target.StreamSinkManagerCtx {
					t.Fatal("successful switch did not select the target")
				}
				source.check(t, 1, 1, false)
				target.check(t, 1, 0, true)
				target.onSample(types.Sample{DeltaUnit: true})
				if counter.Load() != 0 {
					t.Fatal("switched consumer received a delta before a keyframe")
				}
				target.onSample(types.Sample{})
			}
			if counter.Load() != 1 {
				t.Fatal("consumer did not receive the keyframe")
			}
		})
	}
}

func TestPipelineRecreation(t *testing.T) {
	for _, newViewer := range []bool{false, true} {
		t.Run(fmt.Sprintf("new-viewer=%v", newViewer), func(t *testing.T) {
			s := newTestSink(t, codec.VP8())
			selector := streamSelectorNew(codec.VP8(), map[string]*StreamSinkManagerCtx{s.ID(): s.StreamSinkManagerCtx}, []string{s.ID()})
			if newViewer {
				selector.destroyPipelines()
			}
			counter := &sampleCounter{}
			sub := subscribe(t, s, counter)
			if !newViewer {
				selector.destroyPipelines()
			}
			if sub.Stream() != s.StreamSinkManagerCtx || !s.started() {
				t.Fatal("resize lost the logical subscription")
			}
			checkError(t, selector.recreatePipelines())
			if newViewer {
				s.check(t, 1, 0, true)
			} else {
				s.check(t, 2, 1, true)
			}
			s.onSample(types.Sample{})
			if counter.Load() != 1 {
				t.Fatal("subscription stopped delivering after recreation")
			}
		})
	}
}

func TestConcurrentBitrateUpdates(t *testing.T) {
	s := newTestSink(t, codec.Opus())
	s.saveSampleBitrate(time.Unix(2, 0), 1)
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			for range 1000 {
				s.saveSampleBitrate(time.Unix(3, 0), 1)
			}
		})
	}
	wg.Wait()
	s.saveSampleBitrate(time.Unix(4, 0), 0)
	if got := s.Bitrate(); got != 2000 {
		t.Fatalf("bitrate = %d, want 2000", got)
	}
}

func TestBitrateDuringPipelineRestart(t *testing.T) {
	s := newTestSink(t, codec.Opus())
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Go(func() {
		for sec := int64(0); ; sec++ {
			select {
			case <-done:
				return
			default:
				s.saveSampleBitrate(time.Unix(sec, 0), 1)
			}
		}
	})
	defer func() {
		close(done)
		wg.Wait()
	}()
	for range 20 {
		checkError(t, s.createPipeline())
		_ = s.Bitrate()
		s.destroyPipeline()
	}
}
