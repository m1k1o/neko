package types

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type WebSocketMessage struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// UnmarshalJSON normalizes the legacy flat websocket shape into the canonical
// {event, payload} envelope. This lets old clients coexist with the typed SDK
// while keeping all handlers on one payload contract.
func (message *WebSocketMessage) UnmarshalJSON(data []byte) error {
	type wireMessage struct {
		Event   string          `json:"event"`
		Payload json.RawMessage `json:"payload"`
	}

	var wire wireMessage
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	message.Event = wire.Event
	message.Payload = wire.Payload
	trimmed := bytes.TrimSpace(message.Payload)
	if len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	var flat map[string]json.RawMessage
	if err := json.Unmarshal(data, &flat); err != nil {
		return err
	}
	delete(flat, "event")
	delete(flat, "payload")
	if len(flat) == 0 {
		message.Payload = nil
		return nil
	}

	payload, err := json.Marshal(flat)
	if err != nil {
		return err
	}
	message.Payload = payload
	return nil
}

type WebSocketHandler func(Session, WebSocketMessage) bool

type CheckOrigin func(r *http.Request) bool

type WebSocketPeer interface {
	Send(event string, payload any)
	Ping() error
	Destroy(reason string)
}

type WebSocketManager interface {
	Start()
	Shutdown() error
	AddHandler(handler WebSocketHandler)
	Upgrade(checkOrigin CheckOrigin) RouterHandler
}
