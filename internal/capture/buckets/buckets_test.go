package buckets

import (
	"reflect"
	"testing"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
)

func TestBucketsManagerCtx_FindNearestStream(t *testing.T) {
	type fields struct {
		codec   codec.RTPCodec
		streams map[string]types.StreamSinkManager
	}
	type args struct {
		peerBitrate int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   types.StreamSinkManager
	}{
		{
			name: "findNearestStream",
			fields: fields{
				streams: map[string]types.StreamSinkManager{
					"1": mockStreamSink{
						id:      "1",
						bitrate: 500,
					},
					"2": mockStreamSink{
						id:      "2",
						bitrate: 750,
					},
					"3": mockStreamSink{
						id:      "3",
						bitrate: 1000,
					},
					"4": mockStreamSink{
						id:      "4",
						bitrate: 1250,
					},
					"5": mockStreamSink{
						id:      "5",
						bitrate: 1700,
					},
				},
			},
			args: args{
				peerBitrate: 950,
			},
			want: mockStreamSink{
				id:      "2",
				bitrate: 750,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BucketsNew(tt.fields.codec, tt.fields.streams, []string{})

			if got := m.findNearestStream(tt.args.peerBitrate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findNearestStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockStreamSink struct {
	id      string
	bitrate int
	types.StreamSinkManager
}

func (m mockStreamSink) ID() string {
	return m.id
}

func (m mockStreamSink) Bitrate() int {
	return m.bitrate
}
