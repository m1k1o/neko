package types

import (
	"encoding/json"
	"testing"
)

func TestWebSocketMessageOmitsEmptyPayload(t *testing.T) {
	raw, err := json.Marshal(WebSocketMessage{Event: "system/heartbeat"})
	if err != nil {
		t.Fatalf("marshal websocket message: %v", err)
	}

	if got, want := string(raw), `{"event":"system/heartbeat"}`; got != want {
		t.Fatalf("unexpected message: got %s, want %s", got, want)
	}
}

func TestWebSocketMessageUsesCanonicalEnvelope(t *testing.T) {
	raw, err := json.Marshal(WebSocketMessage{
		Event:   "chat/message",
		Payload: json.RawMessage(`{"content":"hello"}`),
	})
	if err != nil {
		t.Fatalf("marshal websocket message: %v", err)
	}

	if got, want := string(raw), `{"event":"chat/message","payload":{"content":"hello"}}`; got != want {
		t.Fatalf("unexpected message: got %s, want %s", got, want)
	}
}

func TestWebSocketMessageRejectsFlatPayload(t *testing.T) {
	var message WebSocketMessage
	if err := json.Unmarshal([]byte(`{"event":"chat/message","content":"hello"}`), &message); err == nil {
		t.Fatal("expected flat websocket message to be rejected")
	}
}

func TestWebSocketMessageRejectsNullPayload(t *testing.T) {
	var message WebSocketMessage
	if err := json.Unmarshal([]byte(`{"event":"system/heartbeat","payload":null}`), &message); err == nil {
		t.Fatal("expected null payload to be rejected")
	}
}
