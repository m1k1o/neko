package gst

import (
	"reflect"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestSampleFromBufferPreservesMetadata(t *testing.T) {
	timestamp := time.Unix(123, 456)
	data := []byte{1, 2, 3}

	tests := []struct {
		name     string
		pts      uint64
		dts      uint64
		duration uint64
		wantPTS  time.Duration
		wantDTS  time.Duration
		wantDur  time.Duration
	}{
		{name: "timestamps", pts: 11, dts: 7, duration: 4, wantPTS: 11, wantDTS: 7, wantDur: 4},
		{name: "clock time none", pts: gstClockTimeNone, dts: gstClockTimeNone, duration: gstClockTimeNone, wantPTS: -1, wantDTS: -1, wantDur: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sampleFromBuffer(data, timestamp, test.pts, test.dts, test.duration, true)
			want := types.Sample{
				Data:      data,
				Timestamp: timestamp,
				PTS:       test.wantPTS,
				DTS:       test.wantDTS,
				Duration:  test.wantDur,
				DeltaUnit: true,
				Length:    len(data),
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("sampleFromBuffer() = %#v, want %#v", got, want)
			}
		})
	}
}
