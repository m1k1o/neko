package media

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

const (
	protocolVersion       byte = 1
	VideoTrack            byte = 1
	AudioTrack            byte = 2
	SampleHeaderSize           = 24
	minSampleBufferSize        = 1 << 10
	maxSampleBufferSize        = 4 << 20
	sampleBufferPoolCount      = 13
)

var sampleBufferPools [sampleBufferPoolCount]sync.Pool

type SampleHeader [SampleHeaderSize]byte

type EncodedSample struct {
	Data      []byte
	poolIndex int
}

type sampleWriter interface {
	WriteSample(byte, types.Sample)
}

type trackConsumer struct {
	peer    sampleWriter
	track   byte
	base    time.Duration
	baseSet bool
}

func newTrackConsumer(peer sampleWriter, track byte) *trackConsumer {
	return &trackConsumer{peer: peer, track: track}
}

func (consumer *trackConsumer) WriteSample(sample types.Sample) {
	timestamp := sample.PTS
	if timestamp < 0 {
		timestamp = sample.DTS
	}
	if timestamp >= 0 {
		if !consumer.baseSet {
			consumer.base = timestamp
			consumer.baseSet = true
		}
		sample.PTS = timestamp - consumer.base
	}
	consumer.peer.WriteSample(consumer.track, sample)
}

func EncodeSample(track byte, sample types.Sample, sequence uint32) *EncodedSample {
	size := SampleHeaderSize + len(sample.Data)
	poolIndex, capacity := sampleBufferPool(size)
	var message *EncodedSample
	if poolIndex >= 0 {
		message, _ = sampleBufferPools[poolIndex].Get().(*EncodedSample)
	}
	if message == nil {
		message = &EncodedSample{Data: make([]byte, capacity)}
	}
	message.Data = message.Data[:size]
	message.poolIndex = poolIndex
	EncodeSampleHeader((*SampleHeader)(message.Data), track, sample, sequence)
	copy(message.Data[SampleHeaderSize:], sample.Data)
	return message
}

func EncodeSampleHeader(header *SampleHeader, track byte, sample types.Sample, sequence uint32) {
	clear(header[:])
	header[0] = protocolVersion
	header[1] = track
	if !sample.DeltaUnit {
		header[2] = 1
	}
	binary.BigEndian.PutUint64(header[4:12], uint64(durationMicroseconds(sample.PTS)))
	binary.BigEndian.PutUint64(header[12:20], uint64(durationMicroseconds(sample.Duration)))
	binary.BigEndian.PutUint32(header[20:24], sequence)
}

func (sample *EncodedSample) Release() {
	if sample == nil || sample.poolIndex < 0 {
		return
	}
	poolIndex := sample.poolIndex
	sample.poolIndex = -1
	sample.Data = sample.Data[:sampleBufferCapacity(poolIndex)]
	sampleBufferPools[poolIndex].Put(sample)
}

func sampleBufferPool(size int) (int, int) {
	if size > maxSampleBufferSize {
		return -1, size
	}
	capacity := minSampleBufferSize
	for poolIndex := 0; poolIndex < sampleBufferPoolCount; poolIndex++ {
		if size <= capacity {
			return poolIndex, capacity
		}
		capacity <<= 1
	}
	return -1, size
}

func sampleBufferCapacity(poolIndex int) int {
	return minSampleBufferSize << poolIndex
}

func durationMicroseconds(duration time.Duration) int64 {
	if duration < 0 {
		return -1
	}
	return duration.Microseconds()
}
