package types

import (
	"encoding/json"
	"net/http"
)

type WebSocketMessage struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type HandlerFunction func(Session, WebSocketMessage) bool

type CheckOrigin func(r *http.Request) bool

type WebSocketPeer interface {
	Send(v interface{}) error
	Destroy()
}

type WebSocketManager interface {
	Start()
	Shutdown() error
	AddHandler(handler HandlerFunction)
	Upgrade(w http.ResponseWriter, r *http.Request, checkOrigin CheckOrigin)
}
