// Package mediaqueue provides bounded queues for real-time media.
//
// When a consumer falls behind, retaining old media increases interactive
// latency. Queue therefore evicts the oldest buffered item before accepting a
// newer one.
package mediaqueue

import (
	"context"
	"sync"
	"sync/atomic"
)

// Stats is a point-in-time snapshot of a Queue.
type Stats struct {
	Capacity int
	Depth    int
	Dropped  uint64
}

// Queue is a close-safe, bounded, latest-item queue. Its channel is
// intentionally never closed: Close notifies blocked readers without racing
// concurrent producers.
type Queue[T any] struct {
	items chan T
	done  chan struct{}
	once  sync.Once

	depth    atomic.Int64
	dropped  atomic.Uint64
	onChange func(Stats)
}

// New creates a queue with a positive, fixed capacity.
func New[T any](capacity int, onChange func(Stats)) *Queue[T] {
	if capacity < 1 {
		panic("media queue capacity must be positive")
	}

	return &Queue[T]{
		items:    make(chan T, capacity),
		done:     make(chan struct{}),
		onChange: onChange,
	}
}

// Push adds item without blocking. If the queue is full, its oldest item is
// discarded so the consumer receives the most recent media.
func (q *Queue[T]) Push(item T) bool {
	select {
	case <-q.done:
		return false
	default:
	}

	select {
	case q.items <- item:
		q.depth.Add(1)
		q.notify()
		return true
	default:
	}

	// There is normally one producer per media queue. The second default makes
	// this safe even if multiple producers temporarily race for an item.
	select {
	case <-q.items:
		q.depth.Add(-1)
		q.dropped.Add(1)
	default:
	}

	select {
	case <-q.done:
		return false
	case q.items <- item:
		q.depth.Add(1)
		q.notify()
		return true
	default:
		return false
	}
}

// Pop waits for the next item or returns false once the queue is closed.
func (q *Queue[T]) Pop() (T, bool) {
	return q.PopContext(context.Background())
}

// PopContext waits for the next item, close, or context cancellation.
func (q *Queue[T]) PopContext(ctx context.Context) (T, bool) {
	// Prefer closure over a buffered item. A closed queue must not deliver
	// stale media even though both cases are selectable.
	select {
	case <-q.done:
		var zero T
		return zero, false
	default:
	}

	select {
	case <-q.done:
		var zero T
		return zero, false
	case <-ctx.Done():
		var zero T
		return zero, false
	case item := <-q.items:
		select {
		case <-q.done:
			var zero T
			return zero, false
		default:
		}
		q.depth.Add(-1)
		q.notify()
		return item, true
	}
}

// Close stops future Push and Pop calls. It is idempotent.
func (q *Queue[T]) Close() {
	q.once.Do(func() {
		close(q.done)
	})
}

// Stats returns a point-in-time queue snapshot.
func (q *Queue[T]) Stats() Stats {
	return Stats{
		Capacity: cap(q.items),
		Depth:    int(q.depth.Load()),
		Dropped:  q.dropped.Load(),
	}
}

func (q *Queue[T]) notify() {
	if q.onChange != nil {
		q.onChange(q.Stats())
	}
}
