package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type WebSocketMessage struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// UnmarshalJSON accepts only the canonical {event, payload} envelope. Keeping
// validation here prevents deprecated flat messages from being silently
// interpreted as empty payloads by individual handlers.
func (m *WebSocketMessage) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	for key := range fields {
		if key != "event" && key != "payload" {
			return errors.New("websocket message contains unsupported top-level fields")
		}
	}

	type envelope WebSocketMessage
	var value envelope
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value.Event == "" {
		return errors.New("websocket message event is required")
	}
	if payload, ok := fields["payload"]; ok && bytes.Equal(bytes.TrimSpace(payload), []byte("null")) {
		return errors.New("websocket message payload must be omitted when empty")
	}

	*m = WebSocketMessage(value)
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
