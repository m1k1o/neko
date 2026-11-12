package capture

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type fakeBroadcastPipeline struct {
	destroyed *atomic.Int32
}

func (p *fakeBroadcastPipeline) Play()    {}
func (p *fakeBroadcastPipeline) Destroy() { p.destroyed.Add(1) }

type testBroadcast struct {
	*BroacastManagerCtx
	empty              atomic.Bool
	created, destroyed atomic.Int32
}

// newTestBroadcast returns a broadcast manager with fake pipelines, watching a room that starts
// empty. The metrics are not registered, so that every test can have its own manager.
func newTestBroadcast(t *testing.T, autostart bool, emptyTimeout time.Duration) *testBroadcast {
	t.Helper()

	b := &testBroadcast{}
	b.BroacastManagerCtx = &BroacastManagerCtx{
		logger:     zerolog.Nop(),
		pipelineFn: func(url string) (string, error) { return url, nil },
		pipelineFactory: func(string) (broadcastPipeline, error) {
			b.created.Add(1)
			return &fakeBroadcastPipeline{&b.destroyed}, nil
		},
		url:              "rtmp://example.com/live",
		started:          autostart,
		emptyTimeout:     emptyTimeout,
		pipelinesCounter: prometheus.NewCounter(prometheus.CounterOpts{Name: "test_pipelines_total"}),
		pipelinesActive:  prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_pipelines_active"}),
	}
	t.Cleanup(b.shutdown)

	if autostart {
		if err := b.createPipeline(); err != nil {
			t.Fatal(err)
		}
	}

	b.empty.Store(true)
	b.watchRoom(b.empty.Load)
	return b
}

func (b *testBroadcast) setEmpty(empty bool) {
	b.empty.Store(empty)
	b.roomChanged()
}

func (b *testBroadcast) counting() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.emptyTimer != nil
}

func (b *testBroadcast) check(t *testing.T, started, counting bool) {
	t.Helper()

	if b.Started() != started || b.counting() != counting {
		t.Fatalf("started, counting = %v, %v, want %v, %v", b.Started(), b.counting(), started, counting)
	}
}

func (b *testBroadcast) waitStopped(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for b.Started() {
		if time.Now().After(deadline) {
			t.Fatal("broadcast was not stopped in an empty room")
		}
		time.Sleep(time.Millisecond)
	}

	if created, destroyed := b.created.Load(), b.destroyed.Load(); created != destroyed {
		t.Fatalf("created %d pipelines, destroyed %d", created, destroyed)
	}
	b.check(t, false, false)
}

func TestBroadcastStopsInEmptyRoom(t *testing.T) {
	b := newTestBroadcast(t, false, 10*time.Millisecond)

	// nothing to stop yet
	b.check(t, false, false)

	if err := b.Start("rtmp://example.com/live"); err != nil {
		t.Fatal(err)
	}
	b.waitStopped(t)
}

func TestBroadcastAutostartStopsInEmptyRoom(t *testing.T) {
	b := newTestBroadcast(t, true, 10*time.Millisecond)
	b.waitStopped(t)
}

func TestBroadcastKeepsRunningWhileRoomIsNotEmpty(t *testing.T) {
	b := newTestBroadcast(t, false, time.Hour)

	b.setEmpty(false)
	if err := b.Start("rtmp://example.com/live"); err != nil {
		t.Fatal(err)
	}
	b.check(t, true, false)

	// the last one leaves
	b.setEmpty(true)
	b.check(t, true, true)

	// someone joins before the timeout
	b.setEmpty(false)
	b.check(t, true, false)

	// stopping cancels the countdown, and a stopped broadcast is not counted down
	b.setEmpty(true)
	b.Stop()
	b.check(t, false, false)
	b.roomChanged()
	b.check(t, false, false)
}

func TestBroadcastCountdownIsNotRestartedWhileRoomStaysEmpty(t *testing.T) {
	b := newTestBroadcast(t, false, time.Hour)

	if err := b.Start("rtmp://example.com/live"); err != nil {
		t.Fatal(err)
	}

	b.mu.Lock()
	timer := b.emptyTimer
	b.mu.Unlock()

	// another session disconnecting must not extend the countdown
	b.setEmpty(true)

	b.mu.Lock()
	defer b.mu.Unlock()
	if timer == nil || b.emptyTimer != timer {
		t.Fatal("countdown was restarted while the room stayed empty")
	}
}

func TestBroadcastNeverStopsWithoutEmptyTimeout(t *testing.T) {
	b := newTestBroadcast(t, true, 0)

	b.setEmpty(true)
	b.check(t, true, false)
}
