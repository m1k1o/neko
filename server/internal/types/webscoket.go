package types

import "net/http"

type WebScoket interface {
	Address() *string
	Send(v interface{}) error
	Destroy() error
}

type WebSocketHandler interface {
	Start() error
	Shutdown() error
	Upgrade(w http.ResponseWriter, r *http.Request) error
}
