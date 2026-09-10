package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/internal/quality"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

func TestApplyVideoProfile(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")

	config := Capture{VideoCodec: codec.VP8(), VideoShowPointer: true}
	if err := config.applyVideoProfile(availableElements("vp8enc")); err != nil {
		t.Fatal(err)
	}
	main := config.VideoPipelines["main"]
	if config.VideoProfile != "balanced" || config.VideoEncoder != quality.EncoderSoftware || len(config.VideoIDs) != 1 || main.Width != "1280" || main.Height != "720" {
		t.Fatalf("unexpected profile config: %+v", config)
	}
}

func TestApplyVideoProfilePreservesDefaultsWhenUnset(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	existing := map[string]struct{}{"custom": {}}
	config := Capture{VideoCodec: codec.VP8(), VideoIDs: []string{"custom"}}
	if err := config.applyVideoProfile(availableElements()); err != nil {
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
	err := config.applyVideoProfile(availableElements("vp8enc"))
	if err == nil || !strings.Contains(err.Error(), "capture.video.pipeline") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyVideoProfileRejectsUnsupportedCodec(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "high")

	config := Capture{VideoCodec: codec.VP9()}
	if err := config.applyVideoProfile(availableElements("vp8enc")); err == nil {
		t.Fatal("profile accepted unsupported codec")
	}
}

func TestApplyVideoProfileFallsBackFromH264ToVP8(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")
	viper.Set("capture.video.encoder", "software")

	config := Capture{VideoCodec: codec.H264()}
	if err := config.applyVideoProfile(availableElements("vp8enc")); err != nil {
		t.Fatal(err)
	}
	if config.VideoCodec.Name != codec.VP8().Name || config.VideoPipelines["main"].GstEncoder != "vp8enc" {
		t.Fatalf("H264 did not fall back to VP8: %+v", config)
	}
}

func TestApplyVideoProfileSupportsAV1Software(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")
	viper.Set("capture.video.codec", "av1")
	viper.Set("capture.video.encoder", "software")

	config := Capture{VideoCodec: codec.AV1()}
	if err := config.applyVideoProfile(availableElements("av1enc")); err != nil {
		t.Fatal(err)
	}
	if config.VideoCodec.Name != codec.AV1().Name || config.VideoPipelines["main"].GstEncoder != "av1enc" {
		t.Fatalf("unexpected AV1 profile config: %+v", config)
	}
}

func TestApplyVideoProfileSupportsH265Software(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")
	viper.Set("capture.video.codec", "h265")
	viper.Set("capture.video.encoder", "software")

	config := Capture{VideoCodec: codec.H265()}
	if err := config.applyVideoProfile(availableElements("x265enc", "h265parse")); err != nil {
		t.Fatal(err)
	}
	if config.VideoCodec.Name != codec.H265().Name || config.VideoPipelines["main"].GstEncoder != "x265enc" {
		t.Fatalf("unexpected H265 profile config: %+v", config)
	}
}

func TestApplyVideoProfileBuildsBrowserCodecFallbackVariants(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "high")
	viper.Set("capture.video.codec", "h265")
	viper.Set("capture.video.encoder", "software")

	config := Capture{VideoCodec: codec.H265()}
	if err := config.applyVideoProfile(availableElements("x265enc", "h265parse", "x264enc", "h264parse", "vp8enc")); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{codec.H265().Name, codec.H264().Name, codec.VP8().Name} {
		variant, ok := config.VideoVariants[name]
		if !ok {
			t.Fatalf("missing browser fallback variant %q: %+v", name, config.VideoVariants)
		}
		if variant.Codec.Name != name || variant.Pipelines["main"].Width != "1920" {
			t.Fatalf("unexpected %s variant: %+v", name, variant)
		}
	}
}

func TestApplyVideoProfileAddsNewCodecHardwareFallback(t *testing.T) {
	tests := []struct {
		name       string
		videoCodec codec.RTPCodec
		hardware   string
		software   string
		parser     string
	}{
		{"av1", codec.AV1(), "nvav1enc", "av1enc", ""},
		{"h265", codec.H265(), "nvh265enc", "x265enc", "h265parse"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			defer viper.Reset()
			viper.Set("capture.video.profile", "balanced")
			viper.Set("capture.video.encoder", "auto")

			available := []string{test.hardware, test.software}
			if test.parser != "" {
				available = append(available, test.parser)
			}
			config := Capture{VideoCodec: test.videoCodec}
			if err := config.applyVideoProfile(availableElements(available...)); err != nil {
				t.Fatal(err)
			}
			if config.VideoPipelines["main"].GstEncoder != test.hardware || len(config.VideoPipelineFallbacks["main"]) != 1 || config.VideoPipelineFallbacks["main"][0].GstEncoder != test.software {
				t.Fatalf("unexpected %s fallback: %+v", test.name, config)
			}
		})
	}
}

func TestApplyVideoProfileAddsSameCodecRuntimeFallback(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")
	viper.Set("capture.video.codec", "h264")

	config := Capture{VideoCodec: codec.H264()}
	if err := config.applyVideoProfile(availableElements("nvautogpuh264enc", "h264parse", "x264enc", "vp8enc")); err != nil {
		t.Fatal(err)
	}
	if config.VideoPipelines["main"].GstEncoder != "nvautogpuh264enc" || len(config.VideoPipelineFallbacks["main"]) != 1 || config.VideoPipelineFallbacks["main"][0].GstEncoder != "x264enc" {
		t.Fatalf("unexpected runtime fallback candidates: %+v", config)
	}
}

func TestApplyVideoProfileFallsBackWhenHardwareDeviceFails(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "balanced")
	viper.Set("capture.video.encoder", "auto")

	config := Capture{VideoCodec: codec.H264()}
	runtimeProbe := func(_ codec.RTPCodec, encoder quality.Encoder, element string) error {
		if encoder == quality.EncoderNVENC && element == "nvautogpuh264enc" {
			return fmt.Errorf("CUDA device unavailable")
		}
		if encoder == quality.EncoderVAAPI {
			return fmt.Errorf("VAAPI driver unavailable")
		}
		return nil
	}
	if err := config.applyVideoProfileWithRuntime(
		availableElements("nvautogpuh264enc", "h264parse", "vah264enc", "x264enc", "vp8enc"),
		runtimeProbe,
	); err != nil {
		t.Fatal(err)
	}
	if config.VideoEncoder != quality.EncoderSoftware || config.VideoPipelines["main"].GstEncoder != "x264enc" {
		t.Fatalf("hardware runtime failure did not select software fallback: %+v", config)
	}
}

func TestApplyVideoProfileBuildsAdaptiveLadderUpToProfile(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.profile", "high")
	viper.Set("capture.video.adaptive", true)

	config := Capture{VideoCodec: codec.VP8()}
	if err := config.applyVideoProfile(availableElements("vp8enc")); err != nil {
		t.Fatal(err)
	}
	if !config.VideoAdaptive || strings.Join(config.VideoIDs, ",") != "low,balanced,high" {
		t.Fatalf("unexpected adaptive video IDs: %+v", config.VideoIDs)
	}
	for _, id := range config.VideoIDs {
		if _, ok := config.VideoPipelines[id]; !ok {
			t.Fatalf("missing adaptive pipeline %q", id)
		}
	}
}

func TestApplyVideoEncoderRequiresProfile(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("capture.video.encoder", "nvenc")

	config := Capture{VideoCodec: codec.H264()}
	if err := config.applyVideoProfile(availableElements("nvh264enc", "h264parse")); err == nil {
		t.Fatal("accepted encoder without a quality profile")
	}
}

func availableElements(available ...string) quality.ElementProbe {
	elements := make(map[string]struct{}, len(available))
	for _, element := range available {
		elements[element] = struct{}{}
	}
	return func(element string) error {
		if _, ok := elements[element]; ok {
			return nil
		}
		return fmt.Errorf("%s unavailable", element)
	}
}
