package codec

import (
	"testing"

	"github.com/pion/webrtc/v4"
)

func TestParseStrIncludesModernVideoCodecs(t *testing.T) {
	for _, name := range []string{"av1", "h264", "h265"} {
		parsed, ok := ParseStr(name)
		if !ok || !parsed.IsVideo() || parsed.Name != name {
			t.Fatalf("ParseStr(%q) = %+v, %v", name, parsed, ok)
		}
	}
}

func TestModernVideoCodecsRegisterWithPion(t *testing.T) {
	for _, videoCodec := range []RTPCodec{AV1(), H264(), H265()} {
		t.Run(videoCodec.Name, func(t *testing.T) {
			engine := &webrtc.MediaEngine{}
			if err := videoCodec.Register(engine); err != nil {
				t.Fatalf("register %s: %v", videoCodec.Name, err)
			}
		})
	}
}

func TestH265CapabilityCoversHighProfile(t *testing.T) {
	if got := H265().Capability.SDPFmtpLine; got != "profile-id=1;level-id=120;tier-flag=0;tx-mode=SRST" {
		t.Fatalf("unexpected H265 capability: %s", got)
	}
}
