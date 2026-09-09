package types

import (
	"encoding/json"
	"testing"
)

func TestWebSocketMessageUnmarshalCanonicalEnvelope(t *testing.T) {
	var message WebSocketMessage
	if err := json.Unmarshal([]byte(`{"event":"signal/offer","payload":{"sdp":"test"}}`), &message); err != nil {
		t.Fatal(err)
	}

	if message.Event != "signal/offer" || string(message.Payload) != `{"sdp":"test"}` {
		t.Fatalf("unexpected message: event=%q payload=%s", message.Event, message.Payload)
	}
}

func TestWebSocketMessageUnmarshalLegacyFlatMessage(t *testing.T) {
	var message WebSocketMessage
	if err := json.Unmarshal([]byte(`{"event":"signal/offer","sdp":"test"}`), &message); err != nil {
		t.Fatal(err)
	}

	if message.Event != "signal/offer" || string(message.Payload) != `{"sdp":"test"}` {
		t.Fatalf("unexpected normalized message: event=%q payload=%s", message.Event, message.Payload)
	}
}

func TestWebSocketMessageUnmarshalEventOnly(t *testing.T) {
	var message WebSocketMessage
	if err := json.Unmarshal([]byte(`{"event":"client/heartbeat"}`), &message); err != nil {
		t.Fatal(err)
	}

	if message.Event != "client/heartbeat" || len(message.Payload) != 0 {
		t.Fatalf("unexpected event-only message: event=%q payload=%s", message.Event, message.Payload)
	}
}
