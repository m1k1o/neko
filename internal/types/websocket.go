package types

import "net/http"

type WebSocket interface {
	Address() string
	Send(v interface{}) error
	Destroy() error
}

type WebSocketManager interface {
	Start()
	Shutdown() error
	Upgrade(w http.ResponseWriter, r *http.Request) error
}
