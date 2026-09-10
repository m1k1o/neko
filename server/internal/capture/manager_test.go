package capture

import (
	"testing"

	"github.com/m1k1o/neko/server/pkg/types/codec"
)

func TestSelectVideoCodecPrefersConfiguredCodecThenCompatibility(t *testing.T) {
	primary := streamSelectorNew(codec.H265(), nil, nil)
	manager := &CaptureManagerCtx{
		video: primary,
		videoVariants: map[string]*StreamSelectorManagerCtx{
			codec.H265().Name: primary,
			codec.H264().Name: streamSelectorNew(codec.H264(), nil, nil),
			codec.VP8().Name:  streamSelectorNew(codec.VP8(), nil, nil),
		},
	}

	selected, ok := manager.SelectVideoCodec([]string{"vp8", "h264"})
	if !ok || selected.Name != codec.H264().Name {
		t.Fatalf("expected H264 compatibility fallback, got %s (ok=%v)", selected.Name, ok)
	}

	selected, ok = manager.SelectVideoCodec([]string{"h265", "h264"})
	if !ok || selected.Name != codec.H265().Name {
		t.Fatalf("expected configured H265 codec, got %s (ok=%v)", selected.Name, ok)
	}
}

func TestSelectVideoCodecRejectsUnsupportedCapabilities(t *testing.T) {
	primary := streamSelectorNew(codec.H265(), nil, nil)
	manager := &CaptureManagerCtx{
		video:         primary,
		videoVariants: map[string]*StreamSelectorManagerCtx{codec.H265().Name: primary},
	}

	if _, ok := manager.SelectVideoCodec([]string{"vp9"}); ok {
		t.Fatal("selected an unavailable codec")
	}

	selected, ok := manager.SelectVideoCodec(nil)
	if !ok || selected.Name != codec.H265().Name {
		t.Fatalf("empty capabilities should preserve configured codec, got %s (ok=%v)", selected.Name, ok)
	}
}
