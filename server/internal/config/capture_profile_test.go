package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/types/codec"
)

func TestApplyVideoProfile(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")

	config := Capture{VideoCodec: codec.VP8(), VideoShowPointer: true}
	if err := config.ApplyVideoProfile(); err != nil {
		t.Fatal(err)
	}
	main := config.VideoPipelines["main"]
	if config.VideoProfile != "balanced" || len(config.VideoIDs) != 1 || main.Width != "1280" || main.Height != "720" {
		t.Fatalf("unexpected profile config: %+v", config)
	}
}

func TestApplyVideoProfilePreservesDefaultsWhenUnset(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	existing := map[string]struct{}{"custom": {}}
	config := Capture{VideoCodec: codec.VP8(), VideoIDs: []string{"custom"}}
	if err := config.ApplyVideoProfile(); err != nil {
		t.Fatal(err)
	}
	if _, ok := existing[config.VideoIDs[0]]; !ok {
		t.Fatalf("unset profile changed video IDs: %v", config.VideoIDs)
	}
}

func TestApplyVideoProfileRejectsCustomPipeline(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "low")
	viper.Set("capture.video.pipeline", "custom ! appsink name=appsink")

	config := Capture{VideoCodec: codec.VP8()}
	err := config.ApplyVideoProfile()
	if err == nil || !strings.Contains(err.Error(), "capture.video.pipeline") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyVideoProfileRejectsUnsupportedCodec(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "high")

	config := Capture{VideoCodec: codec.VP9()}
	if err := config.ApplyVideoProfile(); err == nil {
		t.Fatal("profile accepted unsupported codec")
	}
}
