package types

import (
	"strings"
	"testing"
)

func TestVideoConfigGetPipelineQueues(t *testing.T) {
	screen := ScreenSize{Width: 1920, Height: 1080, Rate: 30}
	leaky := "queue max-size-buffers=1 max-size-time=0 max-size-bytes=0 leaky=downstream"

	tests := []struct {
		name   string
		config VideoConfig
		queues int
	}{
		{"plain", VideoConfig{GstEncoder: "vp8enc"}, 1},
		{"fps", VideoConfig{Fps: "25", GstEncoder: "vp8enc"}, 1},
		{"fps+scale", VideoConfig{Fps: "25", Width: "width / 2", Height: "height / 2", GstEncoder: "vp8enc"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := tt.config.GetPipeline(screen)
			if err != nil {
				t.Fatal(err)
			}
			// every queue in a generated pipeline keeps only the newest frame
			if got := strings.Count(pipeline, "queue"); got != tt.queues {
				t.Errorf("queues = %d, want %d: %s", got, tt.queues, pipeline)
			}
			if got := strings.Count(pipeline, leaky); got != tt.queues {
				t.Errorf("leaky queues = %d, want %d: %s", got, tt.queues, pipeline)
			}
		})
	}
}
