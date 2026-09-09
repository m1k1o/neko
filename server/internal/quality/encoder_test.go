package quality

import (
	"fmt"
	"reflect"
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
