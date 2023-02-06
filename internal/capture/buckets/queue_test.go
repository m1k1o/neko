package buckets

import "testing"

func Queue_normaliseBitrate(t *testing.T) {
	type fields struct {
		queue *queue
	}
	type args struct {
		currentBitrate int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int
	}{
		{
			name: "normaliseBitrate: big drop",
			fields: fields{
				queue: &queue{
					q: []elem{
						{bitrate: 900},
						{bitrate: 750},
						{bitrate: 780},
						{bitrate: 1100},
						{bitrate: 950},
						{bitrate: 700},
						{bitrate: 800},
						{bitrate: 900},
						{bitrate: 1000},
						{bitrate: 1100},
						// avg = 898
					},
				},
			},
			args: args{
				currentBitrate: 350,
			},
			want: []int{816, 700, 537, 350, 350},
		}, {
			name: "normaliseBitrate: small drop",
			fields: fields{
				queue: &queue{
					q: []elem{
						{bitrate: 900},
						{bitrate: 750},
						{bitrate: 780},
						{bitrate: 1100},
						{bitrate: 950},
						{bitrate: 700},
						{bitrate: 800},
						{bitrate: 900},
						{bitrate: 1000},
						{bitrate: 1100},
						// avg = 898
					},
				},
			},
			args: args{
				currentBitrate: 700,
			},
			want: []int{878, 842, 825, 825, 812, 787, 750, 700},
		}, {
			name: "normaliseBitrate",
			fields: fields{
				queue: &queue{
					q: []elem{
						{bitrate: 900},
						{bitrate: 750},
						{bitrate: 780},
						{bitrate: 1100},
						{bitrate: 950},
						{bitrate: 700},
						{bitrate: 800},
						{bitrate: 900},
						{bitrate: 1000},
						{bitrate: 1100},
						// avg = 898
					},
				},
			},
			args: args{
				currentBitrate: 1350,
			},
			want: []int{943, 1003, 1060, 1085},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.queue
			for i := 0; i < len(tt.want); i++ {
				if got := m.normaliseBitrate(tt.args.currentBitrate); got != tt.want[i] {
					t.Errorf("normaliseBitrate() [%d] = %v, want %v", i, got, tt.want[i])
				}
			}
		})
	}
}
