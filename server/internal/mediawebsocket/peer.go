package mediawebsocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/m1k1o/neko/server/internal/media"
	"github.com/m1k1o/neko/server/pkg/types"
)

const (
	// maxUnackedFrames bounds how many video frames may be in flight before the
	// server stops queuing and starts coalescing to the newest frame.
	maxUnackedFrames = 3
	// maxAudioQueue bounds buffered audio. Audio samples are all independently
	// decodable and short, so the oldest is dropped when the client falls behind.
	maxAudioQueue = 32
	// ackStaleAfter releases the unacked gate when a client stops acking
	// entirely, so a lost ack degrades to coalescing instead of a stall.
	ackStaleAfter = 2 * time.Second

	pingInterval = 10 * time.Second
	writeTimeout = 5 * time.Second
)

type peer struct {
	manager    *Manager
	connection *websocket.Conn
	sessionID  string

	requestKeyframe func()

	mu sync.Mutex
	// videoPending holds the newest undelivered video frame. Older frames are
	// released rather than queued, since a stale frame only adds latency.
	videoPending    *media.EncodedSample
	videoPendingSeq uint32
	audioQueue      []*media.EncodedSample
	// sentSeqs tracks sent but unacked video sequence numbers, oldest first.
	sentSeqs  []uint32
	nextSeq   uint32
	acksSeen  bool
	ackedSeq  uint32
	lastAckAt time.Time
	// needKeyframe is set after a video frame is dropped. Delta frames that
	// follow a drop cannot be decoded, so they are discarded until a keyframe.
	needKeyframe bool
	closing      bool

	signal    chan struct{}
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup

	droppedVideo uint64
	droppedAudio uint64
}

func newPeer(manager *Manager, connection *websocket.Conn, sessionID string) *peer {
	return &peer{
		manager:    manager,
		connection: connection,
		sessionID:  sessionID,
		signal:     make(chan struct{}, 1),
		done:       make(chan struct{}),
	}
}

func (peer *peer) SessionID() string {
	return peer.sessionID
}

// WriteSample queues a sample for delivery. It never blocks and never closes the
// connection: a client that cannot keep up is served fewer frames instead.
func (peer *peer) WriteSample(track byte, sample types.Sample) {
	select {
	case <-peer.done:
		return
	default:
	}

	peer.mu.Lock()
	if peer.closing {
		peer.mu.Unlock()
		return
	}

	if track == media.VideoTrack {
		peer.queueVideoLocked(sample)
	} else {
		peer.queueAudioLocked(track, sample)
	}
	peer.mu.Unlock()

	peer.wake()
}

// queueVideoLocked keeps only the newest frame. A keyframe clears a pending
// drop, since it re-establishes a decodable starting point.
func (peer *peer) queueVideoLocked(sample types.Sample) {
	keyframe := !sample.DeltaUnit

	// After a drop the decoder cannot use delta frames until the next keyframe.
	if peer.needKeyframe {
		if !keyframe {
			peer.droppedVideo++
			droppedSamples.WithLabelValues("video", "awaiting_keyframe").Inc()
			return
		}
		peer.needKeyframe = false
	}

	if previous := peer.videoPending; previous != nil {
		// Replacing an undelivered frame breaks the decode chain unless the
		// replacement is itself a keyframe.
		if !keyframe {
			peer.requestRecoveryLocked()
			previous.Release()
			peer.videoPending = nil
			peer.droppedVideo++
			droppedSamples.WithLabelValues("video", "undecodable").Inc()
			return
		}
		previous.Release()
		peer.droppedVideo++
		droppedSamples.WithLabelValues("video", "stale").Inc()
	}

	sequence := peer.nextSeq + 1
	peer.nextSeq = sequence
	peer.videoPending = media.EncodeSample(media.VideoTrack, sample, sequence)
	peer.videoPendingSeq = sequence
}

func (peer *peer) queueAudioLocked(track byte, sample types.Sample) {
	if len(peer.audioQueue) >= maxAudioQueue {
		oldest := peer.audioQueue[0]
		peer.audioQueue = peer.audioQueue[1:]
		oldest.Release()
		peer.droppedAudio++
		droppedSamples.WithLabelValues("audio", "stale").Inc()
	}
	peer.audioQueue = append(peer.audioQueue, media.EncodeSample(track, sample, 0))
	audioQueueDepth.Observe(float64(len(peer.audioQueue)))
}

// requestRecoveryLocked marks the stream as needing a keyframe and asks the
// encoder for one. The encoder throttles repeated requests.
func (peer *peer) requestRecoveryLocked() {
	if peer.needKeyframe {
		return
	}
	peer.needKeyframe = true
	keyframeRequests.Inc()
	if peer.requestKeyframe != nil {
		go peer.requestKeyframe()
	}
}

// videoBlockedLocked reports whether the unacked window is full. The gate only
// applies to clients that have shown they send acks, and is released when acks
// go stale so a lost ack cannot stall delivery permanently.
func (peer *peer) videoBlockedLocked() bool {
	if !peer.acksSeen || len(peer.sentSeqs) < maxUnackedFrames {
		return false
	}
	if time.Since(peer.lastAckAt) > ackStaleAfter {
		peer.sentSeqs = peer.sentSeqs[:0]
		return false
	}
	return true
}

func (peer *peer) onAck(ack media.Ack) {
	peer.mu.Lock()
	peer.acksSeen = true
	peer.lastAckAt = time.Now()
	if ack.Sequence > peer.ackedSeq {
		peer.ackedSeq = ack.Sequence
	}
	// Acks are cumulative: drop every sequence at or below the acked one.
	kept := peer.sentSeqs[:0]
	for _, sequence := range peer.sentSeqs {
		if sequence > ack.Sequence {
			kept = append(kept, sequence)
		}
	}
	peer.sentSeqs = kept
	peer.mu.Unlock()

	peer.wake()
}

func (peer *peer) wake() {
	select {
	case peer.signal <- struct{}{}:
	default:
	}
}

// next returns the next message to write, preferring audio so that a stalled
// video window cannot starve audio playback.
func (peer *peer) next() *media.EncodedSample {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if len(peer.audioQueue) > 0 {
		message := peer.audioQueue[0]
		peer.audioQueue = peer.audioQueue[1:]
		return message
	}

	if peer.videoPending == nil || peer.videoBlockedLocked() {
		return nil
	}

	message := peer.videoPending
	peer.videoPending = nil
	peer.sentSeqs = append(peer.sentSeqs, peer.videoPendingSeq)
	unackedFrames.Observe(float64(len(peer.sentSeqs)))
	// Bound the window for a client that never acks.
	if len(peer.sentSeqs) > maxUnackedFrames {
		peer.sentSeqs = peer.sentSeqs[1:]
	}
	return message
}

func (peer *peer) Close(err error) {
	peer.closeOnce.Do(func() {
		peer.mu.Lock()
		peer.closing = true
		dropped, droppedAudio := peer.droppedVideo, peer.droppedAudio
		peer.mu.Unlock()

		close(peer.done)
		event := peer.manager.logger.Debug()
		if err != nil {
			event = peer.manager.logger.Warn().Err(err)
		}
		event.
			Str("session_id", peer.sessionID).
			Uint64("dropped_video", dropped).
			Uint64("dropped_audio", droppedAudio).
			Msg("closing media websocket")
		_ = peer.connection.Close()
	})
}

func (peer *peer) Run(requestKeyframe func()) {
	peer.mu.Lock()
	peer.requestKeyframe = requestKeyframe
	peer.mu.Unlock()

	peer.wg.Add(1)
	go peer.write()

	for {
		_, data, err := peer.connection.ReadMessage()
		if err != nil {
			peer.Close(nil)
			break
		}
		// Control messages are optional: ignore anything unrecognized so the
		// protocol can grow without breaking older clients.
		control, err := media.DecodeControl(data)
		if err != nil {
			continue
		}
		switch control.Kind {
		case media.ControlAck:
			peer.onAck(media.Ack{Sequence: control.Sequence})
		case media.ControlKeyframeRequest:
			// The client dropped a frame and cannot decode deltas anymore.
			peer.mu.Lock()
			peer.requestRecoveryLocked()
			peer.mu.Unlock()
		}
	}
	peer.wg.Wait()
}

func (peer *peer) write() {
	defer peer.wg.Done()
	defer peer.releaseSamples()
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		// Drain everything currently sendable before waiting again.
		for {
			message := peer.next()
			if message == nil {
				break
			}
			err := peer.connection.WriteMessage(websocket.BinaryMessage, message.Data)
			message.Release()
			if err != nil {
				peer.Close(err)
				return
			}
		}

		select {
		case <-peer.done:
			return
		case <-peer.signal:
		case <-ticker.C:
			if err := peer.connection.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeTimeout)); err != nil {
				peer.Close(err)
				return
			}
		}
	}
}

func (peer *peer) releaseSamples() {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.videoPending != nil {
		peer.videoPending.Release()
		peer.videoPending = nil
	}
	for _, message := range peer.audioQueue {
		message.Release()
	}
	peer.audioQueue = nil
}
