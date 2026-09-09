package quality

import (
	"strings"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

func TestParseProfiles(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		fps         int
		bitrateKbps int
	}{
		{"low", 854, 480, 20, 1000},
		{"balanced", 1280, 720, 30, 2500},
		{"high", 1920, 1080, 30, 4500},
	}
	for _, test := range tests {
		profile, err := Parse(test.name)
		if err != nil {
			t.Fatalf("Parse(%q): %v", test.name, err)
		}
		if profile.Width != test.width || profile.Height != test.height || profile.FPS != test.fps || profile.BitrateKbps != test.bitrateKbps {
			t.Fatalf("Parse(%q) = %+v", test.name, profile)
		}
	}
}

func TestParseRejectsUnknownProfile(t *testing.T) {
	if _, err := Parse("ultra"); err == nil {
		t.Fatal("Parse accepted unknown profile")
	}
}

func TestVideoConfigForVP8(t *testing.T) {
	profile, _ := Parse("balanced")
	config, err := profile.VideoConfig(codec.VP8(), true)
	if err != nil {
		t.Fatal(err)
	}
	if config.GstEncoder != "vp8enc" || config.GstParams["target-bitrate"] != "round(2500 * 1000)" || config.Fps != "30" {
		t.Fatalf("unexpected VP8 config: %+v", config)
	}
	pipeline, err := config.GetPipeline(types.ScreenSize{Width: 1280, Height: 720, Rate: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pipeline, "target-bitrate=2500000") || strings.Contains(pipeline, "e+") {
		t.Fatalf("VP8 bitrate is not rendered as a stable integer: %s", pipeline)
	}
}

func TestVideoConfigForH264(t *testing.T) {
	profile, _ := Parse("low")
	config, err := profile.VideoConfig(codec.H264(), false)
	if err != nil {
		t.Fatal(err)
	}
	if config.GstEncoder != "x264enc" || config.GstParams["bitrate"] != "1000" || config.GstParams["bframes"] != "0" {
		t.Fatalf("unexpected H264 config: %+v", config)
	}
}

func TestVideoConfigRejectsUnsupportedCodec(t *testing.T) {
	profile, _ := Parse("balanced")
	if _, err := profile.VideoConfig(codec.VP9(), true); err == nil {
		t.Fatal("VideoConfig accepted unsupported codec")
	}
}
