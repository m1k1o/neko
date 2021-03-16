package types

import "net/http"

type HandlerFunction func(Session, []byte) bool

type CheckOrigin func(r *http.Request) bool

type WebSocketPeer interface {
	Send(v interface{}) error
	Destroy() error
}

type WebSocketManager interface {
	Start()
	Shutdown() error
	AddHandler(handler HandlerFunction)
	Upgrade(w http.ResponseWriter, r *http.Request, checkOrigin CheckOrigin)
}
