//go:build gstreamer

// Tests that exercise a real GStreamer encoder rather than a fake pipeline.
// They need the gstreamer plugins installed (vp8enc, videotestsrc), so they are
// opt-in:
//
//	go test -tags gstreamer ./internal/capture/
package capture

import (
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

// keyframeCollector records whether each received sample was a keyframe.
type keyframeCollector struct {
	samples chan bool
}

func (c *keyframeCollector) WriteSample(sample types.Sample) {
	select {
	case c.samples <- !sample.DeltaUnit:
	default:
	}
}

// waitForKeyframe consumes samples until a keyframe arrives, returning how many
// delta frames preceded it.
func (c *keyframeCollector) waitForKeyframe(t *testing.T, timeout time.Duration) int {
	t.Helper()
	deadline := time.After(timeout)
	deltas := 0
	for {
		select {
		case keyframe := <-c.samples:
			if keyframe {
				return deltas
			}
			deltas++
		case <-deadline:
			t.Fatalf("no keyframe within %s (%d deltas seen)", timeout, deltas)
			return deltas
		}
	}
}

// RequestKeyframe must produce a keyframe from a real encoder, not merely
// invoke the pipeline call. The fake pipeline used elsewhere cannot show this.
// keyframe-max-dist is set high so a scheduled keyframe cannot be mistaken for
// a forced one.
func TestRequestKeyframeProducesRealKeyframe(t *testing.T) {
	const pipelineStr = "videotestsrc is-live=true pattern=snow ! video/x-raw,framerate=30/1,width=320,height=240 ! " +
		"videoconvert ! vp8enc deadline=1 cpu-used=16 keyframe-max-dist=10000 ! appsink name=appsink"

	manager := streamSinkNew(codec.VP8(), func() (string, error) { return pipelineStr, nil }, "keyframe-real")
	manager.pipelineFactory = func(src string) (sinkPipeline, error) { return gst.CreatePipeline(src) }
	t.Cleanup(manager.shutdown)

	collector := &keyframeCollector{samples: make(chan bool, 256)}
	subscription, err := manager.Subscribe(collector)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = subscription.Close() })

	// Subscribing emits the initial keyframe for the lobby.
	collector.waitForKeyframe(t, 10*time.Second)

	// Let the encoder settle into delta frames.
	settled := false
	deadline := time.After(5 * time.Second)
	for !settled {
		select {
		case keyframe := <-collector.samples:
			if !keyframe {
				settled = true
			}
		case <-deadline:
			t.Fatal("encoder never produced a delta frame")
		}
	}

	// With keyframe-max-dist=10000 at 30fps, a scheduled keyframe is minutes
	// away, so any keyframe now must be the forced one.
	manager.keyframeMu.Lock()
	manager.lastKeyframeReq = time.Time{}
	manager.keyframeMu.Unlock()
	manager.RequestKeyframe()

	deltas := collector.waitForKeyframe(t, 10*time.Second)
	t.Logf("forced keyframe arrived after %d delta frames", deltas)
	if deltas > 60 {
		t.Fatalf("forced keyframe took %d delta frames, too slow to bound latency", deltas)
	}
}

// Without a request, the encoder must not emit a keyframe on its own at this
// setting. This is the control proving the test above measures forcing.
func TestNoSpontaneousKeyframeWithoutRequest(t *testing.T) {
	const pipelineStr = "videotestsrc is-live=true pattern=snow ! video/x-raw,framerate=30/1,width=320,height=240 ! " +
		"videoconvert ! vp8enc deadline=1 cpu-used=16 keyframe-max-dist=10000 ! appsink name=appsink"

	manager := streamSinkNew(codec.VP8(), func() (string, error) { return pipelineStr, nil }, "keyframe-control")
	manager.pipelineFactory = func(src string) (sinkPipeline, error) { return gst.CreatePipeline(src) }
	t.Cleanup(manager.shutdown)

	collector := &keyframeCollector{samples: make(chan bool, 256)}
	subscription, err := manager.Subscribe(collector)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = subscription.Close() })

	collector.waitForKeyframe(t, 10*time.Second)

	// Drain for a while: no further keyframe should appear unprompted.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case keyframe := <-collector.samples:
			if keyframe {
				t.Fatal("encoder emitted an unprompted keyframe, the forcing test would be meaningless")
			}
		case <-deadline:
			return
		}
	}
}
