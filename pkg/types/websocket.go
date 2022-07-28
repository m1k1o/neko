package types

import (
	"encoding/json"
	"net/http"
)

type WebSocketMessage struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
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
