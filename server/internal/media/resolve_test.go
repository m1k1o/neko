package media

import (
	"errors"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type fakeStream struct{ id string }

func (s *fakeStream) ID() string            { return s.id }
func (s *fakeStream) Codec() codec.RTPCodec { return codec.VP8() }
func (s *fakeStream) Bitrate() uint64       { return 0 }
func (s *fakeStream) RequestKeyframe()      {}
func (s *fakeStream) Subscribe(types.SampleConsumer) (types.StreamSubscription, error) {
	return nil, errors.New("not needed")
}

type fakeAudio struct{ fakeStream }

func (s *fakeAudio) Codec() codec.RTPCodec { return codec.Opus() }

type fakeSelector struct{ ids []string }

func (s *fakeSelector) IDs() []string         { return s.ids }
func (s *fakeSelector) Codec() codec.RTPCodec { return codec.VP8() }
func (s *fakeSelector) GetStream(selector types.StreamSelector) (types.EncodedStream, bool) {
	for _, id := range s.ids {
		if id == selector.ID {
			return &fakeStream{id: id}, true
		}
	}
	return nil, false
}

type fakeSource struct{ selector *fakeSelector }

func (s *fakeSource) Audio() types.EncodedStream         { return &fakeAudio{} }
func (s *fakeSource) Video() types.EncodedStreamSelector { return s.selector }

// A client that does not care which quality it gets sends no video id. Resolving that to the
// first configured stream is what lets it connect without discovering the ids first; requiring
// an exact id made every default request fail with "video stream not found".
func TestResolveDefaultsToFirstStream(t *testing.T) {
	manager := &Manager{source: &fakeSource{selector: &fakeSelector{ids: []string{"main", "legacy"}}}}

	video, _, _, _, err := manager.resolve("")
	if err != nil {
		t.Fatalf("empty video id: %v", err)
	}
	if video.ID() != "main" {
		t.Fatalf("default stream = %q, want the first configured one", video.ID())
	}

	// an explicit id still selects exactly that stream
	video, _, _, _, err = manager.resolve("legacy")
	if err != nil || video.ID() != "legacy" {
		t.Fatalf("explicit id = %v, %v", video, err)
	}

	// and an unknown one is still an error rather than a silent fallback
	if _, _, _, _, err := manager.resolve("nope"); err == nil {
		t.Fatal("unknown video id was accepted")
	}
}

func TestResolveWithoutStreams(t *testing.T) {
	manager := &Manager{source: &fakeSource{selector: &fakeSelector{}}}
	if _, _, _, _, err := manager.resolve(""); err == nil {
		t.Fatal("resolving with no configured streams succeeded")
	}
}
