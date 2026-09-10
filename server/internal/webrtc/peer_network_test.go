package webrtc

import "testing"

func TestHasNetworkPressure(t *testing.T) {
	tests := []struct {
		name     string
		queue    float64
		jitter   uint32
		lost     uint32
		baseline uint32
		want     bool
	}{
		{name: "healthy", queue: 0.5, jitter: 1000, lost: 2, baseline: 0, want: false},
		{name: "queue high", queue: videoQueueHighWatermark, want: true},
		{name: "jitter high", jitter: videoJitterHighWatermark, want: true},
		{name: "new loss", lost: 13, baseline: 10, want: true},
		{name: "loss within baseline", lost: 12, baseline: 10, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hasNetworkPressure(test.queue, test.jitter, test.lost, test.baseline); got != test.want {
				t.Fatalf("hasNetworkPressure() = %v, want %v", got, test.want)
			}
		})
	}
}
