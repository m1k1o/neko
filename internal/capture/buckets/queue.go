package buckets

import (
	"math"
	"sync"
	"time"
)

type queue struct {
	sync.Mutex
	q []elem
}

type elem struct {
	created time.Time
	bitrate int
}

func (q *queue) push(v elem) {
	q.Lock()
	defer q.Unlock()

	// if the first element is older than 10 seconds, remove it
	if len(q.q) > 0 && time.Since(q.q[0].created) > 10*time.Second {
		q.q = q.q[1:]
	}
	q.q = append(q.q, v)
}

func (q *queue) len() int {
	q.Lock()
	defer q.Unlock()
	return len(q.q)
}

func (q *queue) avg() int {
	q.Lock()
	defer q.Unlock()
	if len(q.q) == 0 {
		return 0
	}
	sum := 0
	for _, v := range q.q {
		sum += v.bitrate
	}
	return sum / len(q.q)
}

func (q *queue) avgLastN(n int) int {
	if n <= 0 {
		return q.avg()
	}
	q.Lock()
	defer q.Unlock()
	if len(q.q) == 0 {
		return 0
	}
	sum := 0
	for _, v := range q.q[len(q.q)-n:] {
		sum += v.bitrate
	}
	return sum / n
}

func (q *queue) normaliseBitrate(currentBitrate int) int {
	avgBitrate := float64(q.avg())
	histLen := float64(q.len())

	q.push(elem{
		bitrate: currentBitrate,
		created: time.Now(),
	})

	if avgBitrate == 0 || histLen == 0 || currentBitrate == 0 {
		return currentBitrate
	}

	lastN := int(math.Floor(float64(currentBitrate) / avgBitrate * histLen))
	if lastN > q.len() {
		lastN = q.len()
	}

	if lastN == 0 {
		return currentBitrate
	}

	return q.avgLastN(lastN)
}
