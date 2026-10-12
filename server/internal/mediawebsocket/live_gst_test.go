//go:build gstreamer

// End-to-end tests for the backpressure path: a real GStreamer encoder feeding
// a real websocket connection with a genuinely slow client. They need the
// gstreamer plugins installed, so they are opt-in:
//
//	go test -tags gstreamer ./internal/mediawebsocket/
package mediawebsocket

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/internal/media"
	"github.com/m1k1o/neko/server/pkg/gst"
)

// liveEncoder is a real vp8 pipeline. keyframe-max-dist is large so that any
// keyframe observed during a test is a forced one, not a scheduled one.
const liveEncoder = "videotestsrc is-live=true pattern=snow ! video/x-raw,framerate=30/1,width=320,height=240 ! " +
	"videoconvert ! vp8enc deadline=1 cpu-used=16 keyframe-max-dist=10000 ! appsink name=appsink"

type liveFrame struct {
	sequence int
	keyframe bool
	at       time.Time
}

// startLiveStream runs a real encoder into the given peer until cleanup, and
// returns the pipeline so tests can force keyframes through it.
func startLiveStream(t *testing.T, p *peer) gst.Pipeline {
	t.Helper()
	pipeline, err := gst.CreatePipeline(liveEncoder)
	if err != nil {
		t.Fatal(err)
	}
	pipeline.AttachAppsink("appsink")
	pipeline.Play()

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case sample, ok := <-pipeline.Sample():
				if !ok {
					return
				}
				p.WriteSample(media.VideoTrack, sample)
			}
		}
	}()

	t.Cleanup(func() {
		close(done)
		pipeline.Destroy()
		wg.Wait()
	})
	return pipeline
}

// serveLivePeer wires a real websocket server around a peer fed by a real
// encoder, mirroring how media.Manager.Serve runs it in production.
func serveLivePeer(t *testing.T) (*websocket.Conn, *peer) {
	t.Helper()

	var peerRef *peer
	ready := make(chan struct{})
	var served sync.WaitGroup
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		p := newPeer(&Manager{logger: zerolog.Nop()}, connection, "live")
		peerRef = p
		pipeline := startLiveStream(t, p)
		close(ready)

		served.Add(1)
		go func() {
			defer served.Done()
			// The real keyframe path: a client request reaches the encoder.
			p.Run(func() { pipeline.EmitVideoKeyframe() })
		}()
	}))
	t.Cleanup(func() {
		server.Close()
		served.Wait()
	})

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	select {
	case <-ready:
	case <-time.After(10 * time.Second):
		t.Fatal("peer was never served")
	}
	return client, peerRef
}

func readFrame(t *testing.T, client *websocket.Conn) (liveFrame, error) {
	t.Helper()
	_, data, err := client.ReadMessage()
	if err != nil {
		return liveFrame{}, err
	}
	if len(data) < media.SampleHeaderSize {
		t.Fatalf("short frame: %d bytes", len(data))
	}
	return liveFrame{
		sequence: int(binary.BigEndian.Uint32(data[20:24])),
		keyframe: data[2] == 1,
		at:       time.Now(),
	}, nil
}

// A client slower than the encoder must stay connected, and its lag behind the
// encoder must stay bounded instead of growing. This is the behavior the change
// exists to provide: before it, the server closed the connection, and a queuing
// server would leave the client progressively further behind.
func TestSlowClientStaysConnectedAndCurrent(t *testing.T) {
	client, p := serveLivePeer(t)

	// Read far slower than the encoder's 30fps so it must drop frames.
	const reads = 20
	const readInterval = 120 * time.Millisecond

	// Expected steady-state lag: frames encoded during one read interval, plus
	// the unacked window. Lag beyond this means frames are being queued.
	framesPerInterval := int(30 * readInterval / time.Second)
	maxExpectedLag := framesPerInterval + maxUnackedFrames + 2

	var lags []int
	for read := 0; read < reads; read++ {
		time.Sleep(readInterval)
		_ = client.SetReadDeadline(time.Now().Add(10 * time.Second))
		frame, err := readFrame(t, client)
		if err != nil {
			t.Fatalf("connection dropped for a slow client after %d frames: %v", read, err)
		}
		if err := client.WriteMessage(websocket.BinaryMessage, media.EncodeAck(media.Ack{Sequence: uint32(frame.sequence)})); err != nil {
			t.Fatalf("ack failed: %v", err)
		}

		p.mu.Lock()
		produced := int(p.nextSeq)
		p.mu.Unlock()
		lags = append(lags, produced-frame.sequence)
	}

	p.mu.Lock()
	produced := int(p.nextSeq)
	dropped := p.droppedVideo
	p.mu.Unlock()

	// The encoder must have outrun the client, or the test proves nothing.
	if dropped == 0 {
		t.Fatalf("no frames dropped: the client was not slower than the encoder (produced %d)", produced)
	}

	// Lag must be bounded. A queuing server would show it climbing steadily.
	early := lags[1]
	late := lags[len(lags)-1]
	worst := 0
	for _, lag := range lags {
		if lag > worst {
			worst = lag
		}
	}
	t.Logf("produced=%d dropped=%d lag: early=%d late=%d worst=%d (bound %d)",
		produced, dropped, early, late, worst, maxExpectedLag)

	if worst > maxExpectedLag {
		t.Fatalf("lag reached %d frames, beyond the %d expected from coalescing", worst, maxExpectedLag)
	}
	// The decisive check: lag must not grow over time.
	if late > early+maxUnackedFrames {
		t.Fatalf("lag grew from %d to %d frames: the server is queuing, not coalescing", early, late)
	}
}

// Every delivered frame must be decodable: after any gap in sequence numbers
// the next frame must be a keyframe, or the client's decoder is corrupted and
// shows artifacts until the next keyframe arrives.
//
// Two mechanisms guarantee this and are redundant at 30fps, since a forced
// keyframe arrives in roughly one frame time: delta frames are suppressed while
// needKeyframe is set, and the encoder is asked for a keyframe. Removing either
// alone still passes here; removing both produces real corruption. The unit
// tests cover the suppression path directly, which is what protects a slower
// encoder where the request does not arrive in time.
func TestDeliveredFramesRemainDecodable(t *testing.T) {
	client, p := serveLivePeer(t)

	var previous liveFrame
	first := true
	gaps := 0
	// Read slowly enough to force frequent drops, so gaps actually occur.
	for read := 0; read < 25; read++ {
		time.Sleep(150 * time.Millisecond)
		_ = client.SetReadDeadline(time.Now().Add(10 * time.Second))
		frame, err := readFrame(t, client)
		if err != nil {
			t.Fatalf("connection dropped: %v", err)
		}

		if first {
			if !frame.keyframe {
				t.Fatal("the first delivered frame was not a keyframe")
			}
			first = false
		} else if frame.sequence != previous.sequence+1 {
			gaps++
			if !frame.keyframe {
				t.Fatalf("frame %d followed %d with a gap but is a delta: decoder would be corrupted",
					frame.sequence, previous.sequence)
			}
		}

		previous = frame
		if err := client.WriteMessage(websocket.BinaryMessage, media.EncodeAck(media.Ack{Sequence: uint32(frame.sequence)})); err != nil {
			t.Fatalf("ack failed: %v", err)
		}
	}

	p.mu.Lock()
	dropped := p.droppedVideo
	p.mu.Unlock()

	t.Logf("observed %d sequence gaps, %d frames dropped", gaps, dropped)
	// Without gaps the test asserts nothing, so make that failure visible.
	if gaps == 0 {
		t.Fatal("no sequence gaps occurred: the client was not slow enough to exercise recovery")
	}
}

// A client that asks for a keyframe over the wire must get one from the real
// encoder, which is how it recovers after dropping a frame.
func TestClientKeyframeRequestRecoversOverWire(t *testing.T) {
	client, _ := serveLivePeer(t)

	// Drain until the encoder is emitting deltas.
	deadline := time.Now().Add(10 * time.Second)
	for {
		_ = client.SetReadDeadline(time.Now().Add(10 * time.Second))
		frame, err := readFrame(t, client)
		if err != nil {
			t.Fatal(err)
		}
		_ = client.WriteMessage(websocket.BinaryMessage, media.EncodeAck(media.Ack{Sequence: uint32(frame.sequence)}))
		if !frame.keyframe {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("encoder never produced a delta frame")
		}
	}

	// Ask for a keyframe the way a client that dropped a frame would.
	if err := client.WriteMessage(websocket.BinaryMessage, media.EncodeControl(media.ControlKeyframeRequest, 0)); err != nil {
		t.Fatal(err)
	}

	requested := time.Now()
	deadline = time.Now().Add(10 * time.Second)
	for {
		_ = client.SetReadDeadline(time.Now().Add(10 * time.Second))
		frame, err := readFrame(t, client)
		if err != nil {
			t.Fatal(err)
		}
		_ = client.WriteMessage(websocket.BinaryMessage, media.EncodeAck(media.Ack{Sequence: uint32(frame.sequence)}))
		if frame.keyframe {
			t.Logf("keyframe recovered %s after the request", time.Since(requested).Round(time.Millisecond))
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("no keyframe arrived after a client request: a dropped client would stay frozen")
		}
	}
}
