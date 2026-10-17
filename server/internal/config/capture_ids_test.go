package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loadCaptureFromEnvWithIDs mirrors loadCaptureFromEnv but also allows setting
// capture.video.ids directly, so we can check that explicit ids are preserved.
func loadCaptureFromEnvWithIDs(t *testing.T, pipelines string, ids []string) *Capture {
	t.Helper()

	viper.Reset()
	viper.SetEnvPrefix("NEKO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	t.Setenv("NEKO_CAPTURE_VIDEO_PIPELINES", pipelines)

	c := &Capture{}
	cmd := &cobra.Command{}
	if err := c.Init(cmd); err != nil {
		t.Fatal(err)
	}

	if ids != nil {
		viper.Set("capture.video.ids", ids)
	}

	c.Set()
	return c
}

// Regression test for the panic in signalRequest: when video pipelines are
// configured without an explicit capture.video.ids list, VideoIDs must be
// derived from the pipeline keys so the capture manager always has a default
// stream to hand out.
func TestVideoIDsDerivedFromPipelines(t *testing.T) {
	const pipeline = "ximagesrc display-name={display} show-pointer=false ! appsink name=appsink"

	t.Run("multiple pipelines without ids populate sorted ids", func(t *testing.T) {
		c := loadCaptureFromEnvWithIDs(t, `{"high":"`+pipeline+`","medium":"`+pipeline+`","low":"`+pipeline+`"}`, nil)

		want := []string{"high", "low", "medium"}
		if !reflect.DeepEqual(c.VideoIDs, want) {
			t.Fatalf("expected video ids %v derived from pipeline keys, got %v (an empty list makes signalRequest panic)", want, c.VideoIDs)
		}
	})

	t.Run("explicit ids are preserved", func(t *testing.T) {
		c := loadCaptureFromEnvWithIDs(t, `{"high":"`+pipeline+`","low":"`+pipeline+`"}`, []string{"low", "high"})

		want := []string{"low", "high"}
		if !reflect.DeepEqual(c.VideoIDs, want) {
			t.Fatalf("expected explicit video ids %v to be preserved, got %v", want, c.VideoIDs)
		}
	})
}
