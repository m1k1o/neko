package legacy

import (
	"testing"
)

func TestNormalizeClientMessageEnvelope(t *testing.T) {
	got, err := normalizeClientMessage([]byte(`{"event":"signal/offer","payload":{"sdp":"test"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"event":"signal/offer","sdp":"test"}` {
		t.Fatalf("unexpected normalized message: %s", got)
	}
}

func TestNormalizeClientMessageFlat(t *testing.T) {
	raw := []byte(`{"event":"signal/offer","sdp":"test"}`)
	got, err := normalizeClientMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"event":"signal/offer","sdp":"test"}` {
		t.Fatalf("unexpected normalized message: %s", got)
	}
}
