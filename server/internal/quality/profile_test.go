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
	config, err := profile.VideoConfig(codec.VP8(), EncoderSoftware, "vp8enc", true)
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
	config, err := profile.VideoConfig(codec.H264(), EncoderSoftware, "x264enc", false)
	if err != nil {
		t.Fatal(err)
	}
	if config.GstEncoder != "x264enc" || config.GstParams["bitrate"] != "1000" || config.GstParams["bframes"] != "0" {
		t.Fatalf("unexpected H264 config: %+v", config)
	}
}

func TestVideoConfigForH264HardwareEncoders(t *testing.T) {
	profile, _ := Parse("high")
	for _, test := range []struct {
		encoder Encoder
		element string
	}{
		{EncoderVAAPI, "vah264enc"},
		{EncoderVAAPI, "vah264lpenc"},
		{EncoderNVENC, "nvh264enc"},
		{EncoderNVENC, "nvautogpuh264enc"},
	} {
		config, err := profile.VideoConfig(codec.H264(), test.encoder, test.element, false)
		if err != nil {
			t.Fatalf("%s/%s: %v", test.encoder, test.element, err)
		}
		pipeline, err := config.GetPipeline(types.ScreenSize{Width: 1920, Height: 1080, Rate: 30})
		if err != nil {
			t.Fatalf("%s/%s pipeline: %v", test.encoder, test.element, err)
		}
		if !strings.Contains(pipeline, "! "+test.element+" name=encoder") || !strings.Contains(pipeline, "! h264parse") {
			t.Fatalf("unexpected %s pipeline: %s", test.encoder, pipeline)
		}
	}
}

func TestVideoConfigForH265Encoders(t *testing.T) {
	profile, _ := Parse("balanced")
	tests := []struct {
		encoder Encoder
		element string
	}{
		{EncoderSoftware, "x265enc"},
		{EncoderVAAPI, "vah265enc"},
		{EncoderVAAPI, "vah265lpenc"},
		{EncoderNVENC, "nvh265enc"},
		{EncoderNVENC, "nvautogpuh265enc"},
	}
	for _, test := range tests {
		config, err := profile.VideoConfig(codec.H265(), test.encoder, test.element, false)
		if err != nil {
			t.Fatalf("%s/%s: %v", test.encoder, test.element, err)
		}
		pipeline, err := config.GetPipeline(types.ScreenSize{Width: 1280, Height: 720, Rate: 30})
		if err != nil {
			t.Fatalf("%s/%s pipeline: %v", test.encoder, test.element, err)
		}
		if !strings.Contains(pipeline, "! "+test.element+" name=encoder") || !strings.Contains(pipeline, "! h265parse") {
			t.Fatalf("unexpected %s pipeline: %s", test.encoder, pipeline)
		}
	}
}

func TestVideoConfigForAV1Encoders(t *testing.T) {
	profile, _ := Parse("high")
	tests := []struct {
		encoder Encoder
		element string
	}{
		{EncoderSoftware, "av1enc"},
		{EncoderSoftware, "svtav1enc"},
		{EncoderVAAPI, "vaav1enc"},
		{EncoderNVENC, "nvav1enc"},
		{EncoderNVENC, "nvautogpuav1enc"},
	}
	for _, test := range tests {
		config, err := profile.VideoConfig(codec.AV1(), test.encoder, test.element, false)
		if err != nil {
			t.Fatalf("%s/%s: %v", test.encoder, test.element, err)
		}
		pipeline, err := config.GetPipeline(types.ScreenSize{Width: 1920, Height: 1080, Rate: 30})
		if err != nil {
			t.Fatalf("%s/%s pipeline: %v", test.encoder, test.element, err)
		}
		if !strings.Contains(pipeline, "! "+test.element+" name=encoder") || !strings.Contains(pipeline, "video/x-av1,stream-format=obu-stream") {
			t.Fatalf("unexpected %s pipeline: %s", test.encoder, pipeline)
		}
	}
}

func TestVideoConfigRejectsUnsupportedCodec(t *testing.T) {
	profile, _ := Parse("balanced")
	if _, err := profile.VideoConfig(codec.VP9(), EncoderSoftware, "vp9enc", true); err == nil {
		t.Fatal("VideoConfig accepted unsupported codec")
	}
}
