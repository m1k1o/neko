package config

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loadCaptureFromEnv mirrors the production config flow: env prefix setup,
// flag registration via Init, then value population via Set.
func loadCaptureFromEnv(t *testing.T, pipelines string) *Capture {
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
	c.Set()
	return c
}

func TestVideoPipelinesEnv(t *testing.T) {
	const pipeline = "ximagesrc display-name={display} show-pointer=false ! appsink name=appsink"

	t.Run("object entries with gst_pipeline keep working", func(t *testing.T) {
		c := loadCaptureFromEnv(t, `{"high":{"gst_pipeline":"`+pipeline+`"}}`)

		cfg, ok := c.VideoPipelines["high"]
		if !ok {
			t.Fatalf("pipeline entry missing, got %v", c.VideoPipelines)
		}
		if cfg.GstPipeline != pipeline {
			t.Errorf("expected pipeline %q, got %q", pipeline, cfg.GstPipeline)
		}
	})

	t.Run("bare pipeline string entries are accepted", func(t *testing.T) {
		c := loadCaptureFromEnv(t, `{"high":"`+pipeline+`"}`)

		cfg, ok := c.VideoPipelines["high"]
		if !ok {
			t.Fatalf("bare pipeline string entry was dropped and neko fell back to the default vp8 pipeline, got %v", c.VideoPipelines)
		}
		if cfg.GstPipeline != pipeline {
			t.Errorf("expected pipeline %q, got %q", pipeline, cfg.GstPipeline)
		}
	})
}
