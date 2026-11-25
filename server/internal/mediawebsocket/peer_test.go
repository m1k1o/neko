package mediawebsocket

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/internal/media"
	"github.com/m1k1o/neko/server/pkg/types"
)

func testPeer(t *testing.T) *peer {
	t.Helper()
	p := newPeer(&Manager{logger: zerolog.Nop()}, nil, "session")
	t.Cleanup(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.videoPending != nil {
			p.videoPending.Release()
			p.videoPending = nil
		}
		for _, message := range p.audioQueue {
			message.Release()
		}
		p.audioQueue = nil
	})
	return p
}

func video(keyframe bool) types.Sample {
	return types.Sample{Data: []byte{1, 2, 3}, DeltaUnit: !keyframe}
}

// take drains the next sendable message and reports its sequence number.
func take(t *testing.T, p *peer) (uint32, bool) {
	t.Helper()
	message := p.next()
	if message == nil {
		return 0, false
	}
	defer message.Release()
	if message.Data[1] != media.VideoTrack {
		t.Fatalf("track = %d, want video", message.Data[1])
	}
	return binary.BigEndian.Uint32(message.Data[20:24]), true
}

func TestVideoCoalescesToNewestKeyframe(t *testing.T) {
	p := testPeer(t)

	p.WriteSample(media.VideoTrack, video(true))
	p.WriteSample(media.VideoTrack, video(true))

	sequence, ok := take(t, p)
	if !ok {
		t.Fatal("no frame available")
	}
	if sequence != 2 {
		t.Fatalf("delivered sequence = %d, want the newest frame", sequence)
	}
	if _, ok := take(t, p); ok {
		t.Fatal("stale frame was queued behind the newest one")
	}
}

// Dropping a frame invalidates the decode chain, so deltas must be withheld
// until a keyframe re-establishes a decodable starting point.
func TestDroppedFrameWithholdsDeltasUntilKeyframe(t *testing.T) {
	p := testPeer(t)
	var requested int
	var mu sync.Mutex
	p.requestKeyframe = func() {
		mu.Lock()
		requested++
		mu.Unlock()
	}

	// Two undelivered frames where the newest is a delta: both are unusable.
	p.WriteSample(media.VideoTrack, video(true))
	p.WriteSample(media.VideoTrack, video(false))

	if _, ok := take(t, p); ok {
		t.Fatal("a frame was delivered after the decode chain broke")
	}

	// Deltas stay suppressed while waiting for recovery.
	p.WriteSample(media.VideoTrack, video(false))
	if _, ok := take(t, p); ok {
		t.Fatal("a delta frame was delivered before a keyframe")
	}

	// A keyframe restores delivery.
	p.WriteSample(media.VideoTrack, video(true))
	if _, ok := take(t, p); !ok {
		t.Fatal("keyframe was not delivered after recovery")
	}

	// Wait for the asynchronous keyframe request.
	deadline := time.Now().Add(time.Second)
	for {
		mu.Lock()
		count := requested
		mu.Unlock()
		if count > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no keyframe was requested after dropping a frame")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestUnackedWindowBlocksThenReleases(t *testing.T) {
	p := testPeer(t)

	// The window only applies once the client has proven it acks.
	p.onAck(media.Ack{Sequence: 0})

	var last uint32
	for sent := 0; sent < maxUnackedFrames; sent++ {
		p.WriteSample(media.VideoTrack, video(true))
		sequence, ok := take(t, p)
		if !ok {
			t.Fatalf("frame %d was not delivered", sent)
		}
		last = sequence
	}

	p.WriteSample(media.VideoTrack, video(true))
	if _, ok := take(t, p); ok {
		t.Fatal("a frame was delivered with a full unacked window")
	}

	p.onAck(media.Ack{Sequence: last})
	if _, ok := take(t, p); !ok {
		t.Fatal("acking did not release the window")
	}
}

// A client that never acks must not be starved by the window.
func TestNonAckingClientKeepsReceiving(t *testing.T) {
	p := testPeer(t)

	for sent := 0; sent < maxUnackedFrames*3; sent++ {
		p.WriteSample(media.VideoTrack, video(true))
		if _, ok := take(t, p); !ok {
			t.Fatalf("frame %d was withheld from a client that never acks", sent)
		}
	}
}

// A client that acked once and then stopped must recover rather than stall.
func TestStaleAcksReleaseWindow(t *testing.T) {
	p := testPeer(t)
	p.onAck(media.Ack{Sequence: 0})

	for sent := 0; sent < maxUnackedFrames; sent++ {
		p.WriteSample(media.VideoTrack, video(true))
		take(t, p)
	}

	p.WriteSample(media.VideoTrack, video(true))
	if _, ok := take(t, p); ok {
		t.Fatal("window was not enforced")
	}

	p.mu.Lock()
	p.lastAckAt = time.Now().Add(-2 * ackStaleAfter)
	p.mu.Unlock()

	if _, ok := take(t, p); !ok {
		t.Fatal("stale acks did not release the window")
	}
}

// Audio must keep flowing while the video window is saturated.
func TestAudioIsNotBlockedByVideoWindow(t *testing.T) {
	p := testPeer(t)
	p.onAck(media.Ack{Sequence: 0})

	for sent := 0; sent < maxUnackedFrames; sent++ {
		p.WriteSample(media.VideoTrack, video(true))
		take(t, p)
	}
	p.WriteSample(media.VideoTrack, video(true))

	p.WriteSample(media.AudioTrack, types.Sample{Data: []byte{9}})
	message := p.next()
	if message == nil {
		t.Fatal("audio was blocked by a saturated video window")
	}
	defer message.Release()
	if message.Data[1] != media.AudioTrack {
		t.Fatalf("track = %d, want audio", message.Data[1])
	}
}

// A client that dropped a frame itself can ask for a keyframe, which must reach
// the encoder and suppress deltas until it arrives.
func TestClientKeyframeRequestSuppressesDeltas(t *testing.T) {
	p := testPeer(t)
	requested := make(chan struct{}, 4)
	p.requestKeyframe = func() { requested <- struct{}{} }

	p.mu.Lock()
	p.requestRecoveryLocked()
	p.mu.Unlock()

	select {
	case <-requested:
	case <-time.After(time.Second):
		t.Fatal("a client keyframe request did not reach the encoder")
	}

	p.WriteSample(media.VideoTrack, video(false))
	if _, ok := take(t, p); ok {
		t.Fatal("a delta was delivered while waiting for the requested keyframe")
	}

	p.WriteSample(media.VideoTrack, video(true))
	if _, ok := take(t, p); !ok {
		t.Fatal("the keyframe was not delivered after the request")
	}
}

// The read loop must route real client control messages, not just the internal
// helpers: this covers the wiring from the wire format to peer state.
func TestReadLoopRoutesControlMessages(t *testing.T) {
	requested := make(chan struct{}, 4)

	var served sync.WaitGroup
	var peerRef atomic.Pointer[peer]
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		p := newPeer(&Manager{logger: zerolog.Nop()}, connection, "session")
		peerRef.Store(p)
		served.Add(1)
		go func() {
			defer served.Done()
			p.Run(func() { requested <- struct{}{} })
		}()
	}))
	defer server.Close()

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the server to construct the peer.
	deadline := time.Now().Add(2 * time.Second)
	var p *peer
	for p == nil {
		if time.Now().After(deadline) {
			t.Fatal("peer was never created")
		}
		p = peerRef.Load()
		time.Sleep(time.Millisecond)
	}

	// A keyframe request must reach the encoder through the read loop.
	if err := client.WriteMessage(websocket.BinaryMessage, media.EncodeControl(media.ControlKeyframeRequest, 0)); err != nil {
		t.Fatal(err)
	}
	select {
	case <-requested:
	case <-time.After(2 * time.Second):
		t.Fatal("the read loop did not route a keyframe request to the encoder")
	}

	// An ack must be applied to the peer's window through the read loop.
	if err := client.WriteMessage(websocket.BinaryMessage, media.EncodeAck(media.Ack{Sequence: 11})); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for {
		p.mu.Lock()
		seen := p.acksSeen
		p.mu.Unlock()
		if seen {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the read loop did not apply a client ack")
		}
		time.Sleep(time.Millisecond)
	}

	// A malformed message must be ignored rather than closing the connection.
	if err := client.WriteMessage(websocket.BinaryMessage, []byte{9, 9}); err != nil {
		t.Fatal(err)
	}
	// Writes are buffered locally, so confirm the peer is still live by
	// observing that a later ack is still applied.
	if err := client.WriteMessage(websocket.BinaryMessage, media.EncodeAck(media.Ack{Sequence: 12})); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for {
		p.mu.Lock()
		closing := p.closing
		p.mu.Unlock()
		if closing {
			t.Fatal("a malformed control message closed the connection")
		}

		p.mu.Lock()
		acked := p.ackedSeq
		p.mu.Unlock()
		if acked == 12 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the peer stopped processing control messages after a malformed one")
		}
		time.Sleep(time.Millisecond)
	}

	_ = client.Close()
	served.Wait()
}

// Drops must be attributed by reason: "stale" is the window working as
// designed, while "undecodable" means the client sees a visible freeze.
func TestDropMetricsDistinguishReasons(t *testing.T) {
	stale := testutil.ToFloat64(droppedSamples.WithLabelValues("video", "stale"))
	undecodable := testutil.ToFloat64(droppedSamples.WithLabelValues("video", "undecodable"))
	audio := testutil.ToFloat64(droppedSamples.WithLabelValues("audio", "stale"))

	p := testPeer(t)
	// Two keyframes: the older one is stale and safe to discard.
	p.WriteSample(media.VideoTrack, video(true))
	p.WriteSample(media.VideoTrack, video(true))
	if got := testutil.ToFloat64(droppedSamples.WithLabelValues("video", "stale")) - stale; got != 1 {
		t.Fatalf("stale video drops = %f, want 1", got)
	}

	// Replacing a pending frame with a delta breaks the decode chain.
	p.WriteSample(media.VideoTrack, video(false))
	if got := testutil.ToFloat64(droppedSamples.WithLabelValues("video", "undecodable")) - undecodable; got != 1 {
		t.Fatalf("undecodable video drops = %f, want 1", got)
	}

	q := testPeer(t)
	for sent := 0; sent < maxAudioQueue+2; sent++ {
		q.WriteSample(media.AudioTrack, types.Sample{Data: []byte{1}})
	}
	if got := testutil.ToFloat64(droppedSamples.WithLabelValues("audio", "stale")) - audio; got != 2 {
		t.Fatalf("audio drops = %f, want 2", got)
	}
}

func TestAudioQueueDropsOldestWhenFull(t *testing.T) {
	p := testPeer(t)

	for sent := 0; sent < maxAudioQueue+5; sent++ {
		p.WriteSample(media.AudioTrack, types.Sample{Data: []byte{byte(sent)}})
	}

	p.mu.Lock()
	queued := len(p.audioQueue)
	dropped := p.droppedAudio
	p.mu.Unlock()

	if queued != maxAudioQueue {
		t.Fatalf("queued audio = %d, want %d", queued, maxAudioQueue)
	}
	if dropped != 5 {
		t.Fatalf("dropped audio = %d, want 5", dropped)
	}
}
