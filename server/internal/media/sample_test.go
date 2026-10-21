package media

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

type sampleRecorder struct {
	samples [][]byte
}

func (recorder *sampleRecorder) WriteSample(track byte, sample types.Sample) {
	encoded := EncodeSample(track, sample, 0)
	recorder.samples = append(recorder.samples, append([]byte(nil), encoded.Data...))
	encoded.Release()
}

func TestEncodeSample(t *testing.T) {
	sample := types.Sample{
		Data:      []byte{1, 2, 3},
		PTS:       1234567 * time.Nanosecond,
		Duration:  -1,
		DeltaUnit: false,
	}

	message := EncodeSample(VideoTrack, sample, 42)
	defer message.Release()
	if len(message.Data) != SampleHeaderSize+len(sample.Data) {
		t.Fatalf("message length = %d", len(message.Data))
	}
	if message.Data[0] != protocolVersion || message.Data[1] != VideoTrack || message.Data[2] != 1 {
		t.Fatalf("header = %v", message.Data[:4])
	}
	if got := int64(binary.BigEndian.Uint64(message.Data[4:12])); got != 1234 {
		t.Fatalf("timestamp = %d", got)
	}
	if got := int64(binary.BigEndian.Uint64(message.Data[12:20])); got != -1 {
		t.Fatalf("duration = %d", got)
	}
	if got := binary.BigEndian.Uint32(message.Data[20:24]); got != 42 {
		t.Fatalf("sequence = %d", got)
	}
	for index, value := range sample.Data {
		if message.Data[SampleHeaderSize+index] != value {
			t.Fatalf("payload = %v", message.Data[SampleHeaderSize:])
		}
	}
}

func TestEncodeSampleClearsPooledHeader(t *testing.T) {
	keyframe := EncodeSample(VideoTrack, types.Sample{Data: []byte{1}}, 7)
	if keyframe.Data[2] != 1 {
		t.Fatal("keyframe flag is not set")
	}
	keyframe.Release()

	delta := EncodeSample(VideoTrack, types.Sample{Data: []byte{2}, DeltaUnit: true}, 8)
	defer delta.Release()
	if delta.Data[2] != 0 {
		t.Fatal("stale keyframe flag was retained")
	}
	if got := binary.BigEndian.Uint32(delta.Data[20:24]); got != 8 {
		t.Fatalf("stale sequence was retained = %d", got)
	}
}

func TestSampleBufferPoolBounds(t *testing.T) {
	if poolIndex, capacity := sampleBufferPool(maxSampleBufferSize); poolIndex < 0 || capacity != maxSampleBufferSize {
		t.Fatalf("maximum pooled buffer = (%d, %d)", poolIndex, capacity)
	}
	if poolIndex, capacity := sampleBufferPool(maxSampleBufferSize + 1); poolIndex != -1 || capacity != maxSampleBufferSize+1 {
		t.Fatalf("oversized buffer = (%d, %d)", poolIndex, capacity)
	}
}

func TestEncodeSampleReusesBuffer(t *testing.T) {
	sample := types.Sample{Data: make([]byte, 4096)}
	encoded := EncodeSample(VideoTrack, sample, 0)
	encoded.Release()

	allocations := testing.AllocsPerRun(100, func() {
		encoded := EncodeSample(VideoTrack, sample, 0)
		encoded.Release()
	})
	if allocations != 0 {
		t.Fatalf("allocations per encoding = %f", allocations)
	}
}

func TestTrackConsumerNormalizesTimestamps(t *testing.T) {
	recorder := &sampleRecorder{}
	consumer := newTrackConsumer(recorder, AudioTrack)
	consumer.WriteSample(types.Sample{PTS: 5 * time.Second, Duration: 20 * time.Millisecond})
	consumer.WriteSample(types.Sample{PTS: 5020 * time.Millisecond, Duration: 20 * time.Millisecond})

	if got := int64(binary.BigEndian.Uint64(recorder.samples[0][4:12])); got != 0 {
		t.Fatalf("first timestamp = %d", got)
	}
	if got := int64(binary.BigEndian.Uint64(recorder.samples[1][4:12])); got != 20000 {
		t.Fatalf("second timestamp = %d", got)
	}
}
