package quality

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types/codec"
)

func elementProbe(available ...string) ElementProbe {
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

func TestResolveEncoderAutoPrefersNVENC(t *testing.T) {
	selection, err := ResolveEncoder(codec.H264(), EncoderAuto, elementProbe("nvautogpuh264enc", "h264parse", "x264enc", "vp8enc"))
	if err != nil {
		t.Fatal(err)
	}
	if selection.Encoder != EncoderNVENC || selection.Element != "nvautogpuh264enc" || selection.Codec.Name != codec.H264().Name {
		t.Fatalf("unexpected selection: %+v", selection)
	}
}

func TestParseEncoderDefaultsToAuto(t *testing.T) {
	encoder, err := ParseEncoder("")
	if err != nil || encoder != EncoderAuto {
		t.Fatalf("unexpected default encoder: %q, %v", encoder, err)
	}
}

func TestResolveEncoderNVENCFallsBackToX264(t *testing.T) {
	selection, err := ResolveEncoder(codec.H264(), EncoderNVENC, elementProbe("x264enc", "h264parse", "vp8enc"))
	if err != nil {
		t.Fatal(err)
	}
	if selection.Encoder != EncoderSoftware || selection.Element != "x264enc" || !reflect.DeepEqual(selection.Unavailable, []string{"nvautogpuh264enc", "nvh264enc"}) {
		t.Fatalf("unexpected selection: %+v", selection)
	}
}

func TestResolveEncoderFallsBackWhenHardwareDeviceFails(t *testing.T) {
	selection, err := ResolveEncoderWithRuntime(
		codec.H264(),
		EncoderAuto,
		elementProbe("nvautogpuh264enc", "h264parse", "vah264enc", "x264enc", "vp8enc"),
		func(_ codec.RTPCodec, encoder Encoder, element string) error {
			if encoder == EncoderNVENC && element == "nvautogpuh264enc" {
				return fmt.Errorf("CUDA device unavailable")
			}
			if encoder == EncoderVAAPI {
				return fmt.Errorf("VAAPI driver unavailable")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Encoder != EncoderSoftware || selection.Element != "x264enc" {
		t.Fatalf("hardware runtime failure did not fall back to software: %+v", selection)
	}
	if !containsDiagnostic(selection.Unavailable, "CUDA device unavailable") || !containsDiagnostic(selection.Unavailable, "VAAPI driver unavailable") {
		t.Fatalf("runtime failure diagnostics were not retained: %+v", selection.Unavailable)
	}
}

func containsDiagnostic(diagnostics []string, value string) bool {
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic, value) {
			return true
		}
	}
	return false
}

func TestResolveEncoderRuntimeProbeOnlyChecksHardware(t *testing.T) {
	called := 0
	selection, err := ResolveEncoderWithRuntime(
		codec.H264(),
		EncoderSoftware,
		elementProbe("x264enc", "h264parse", "vp8enc"),
		func(_ codec.RTPCodec, encoder Encoder, element string) error {
			called++
			return fmt.Errorf("unexpected runtime probe for %s/%s", encoder, element)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Encoder != EncoderSoftware || selection.Element != "x264enc" || called != 0 {
		t.Fatalf("software selection unexpectedly used runtime probe: %+v calls=%d", selection, called)
	}
}

func TestResolveEncoderTriesHardwareAlternativeAfterRuntimeFailure(t *testing.T) {
	selection, err := ResolveEncoderWithRuntime(
		codec.H264(),
		EncoderNVENC,
		elementProbe("nvautogpuh264enc", "nvh264enc", "h264parse"),
		func(_ codec.RTPCodec, _ Encoder, element string) error {
			if element == "nvautogpuh264enc" {
				return fmt.Errorf("CUDA device unavailable")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Encoder != EncoderNVENC || selection.Element != "nvh264enc" {
		t.Fatalf("did not try the second hardware element: %+v", selection)
	}
	if !containsDiagnostic(selection.Unavailable, "CUDA device unavailable") {
		t.Fatalf("missing first hardware diagnostic: %+v", selection.Unavailable)
	}
}

func TestResolveEncoderH264FallsBackToVP8(t *testing.T) {
	selection, err := ResolveEncoder(codec.H264(), EncoderSoftware, elementProbe("vp8enc"))
	if err != nil {
		t.Fatal(err)
	}
	if selection.Codec.Name != codec.VP8().Name || selection.Element != "vp8enc" {
		t.Fatalf("unexpected selection: %+v", selection)
	}
}

func TestResolveEncoderRejectsHardwareVP8(t *testing.T) {
	if _, err := ResolveEncoder(codec.VP8(), EncoderVAAPI, elementProbe("vp8enc")); err == nil {
		t.Fatal("accepted VAAPI for VP8")
	}
}

func TestResolveEncoderFailsWithoutFallback(t *testing.T) {
	if _, err := ResolveEncoder(codec.H264(), EncoderAuto, elementProbe()); err == nil {
		t.Fatal("resolved encoder without any available element")
	}
}

func TestResolveEncoderAutoSupportsAV1HardwareMatrix(t *testing.T) {
	tests := []struct {
		name    string
		element string
	}{
		{"nvidia-auto", "nvautogpuav1enc"},
		{"nvidia-cuda", "nvav1enc"},
		{"vaapi", "vaav1enc"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selection, err := ResolveEncoder(codec.AV1(), EncoderAuto, elementProbe(test.element))
			if err != nil {
				t.Fatal(err)
			}
			if selection.Codec.Name != codec.AV1().Name || selection.Element != test.element || selection.Encoder == EncoderSoftware {
				t.Fatalf("unexpected AV1 selection: %+v", selection)
			}
		})
	}
}

func TestResolveEncoderAutoSupportsH265HardwareMatrix(t *testing.T) {
	tests := []struct {
		name    string
		element string
	}{
		{"nvidia-auto", "nvautogpuh265enc"},
		{"nvidia-cuda", "nvh265enc"},
		{"vaapi", "vah265enc"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selection, err := ResolveEncoder(codec.H265(), EncoderAuto, elementProbe(test.element, "h265parse"))
			if err != nil {
				t.Fatal(err)
			}
			if selection.Codec.Name != codec.H265().Name || selection.Element != test.element || selection.Encoder == EncoderSoftware {
				t.Fatalf("unexpected H265 selection: %+v", selection)
			}
		})
	}
}

func TestResolveEncoderNewCodecFallsBackToH264(t *testing.T) {
	for _, videoCodec := range []codec.RTPCodec{codec.AV1(), codec.H265()} {
		selection, err := ResolveEncoder(videoCodec, EncoderSoftware, elementProbe("x264enc", "h264parse"))
		if err != nil {
			t.Fatalf("%s: %v", videoCodec.Name, err)
		}
		if selection.Codec.Name != codec.H264().Name || selection.Element != "x264enc" {
			t.Fatalf("%s did not fall back to H264: %+v", videoCodec.Name, selection)
		}
	}
}
