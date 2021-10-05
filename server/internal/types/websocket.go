package types

import "net/http"

type Stats struct {
	Connections uint32    `json:"connections"`
	Host        string    `json:"host"`
	Members     []*Member `json:"members"`
}

type WebSocket interface {
	Address() string
	Send(v interface{}) error
	Destroy() error
}

type WebSocketHandler interface {
	Start()
	Shutdown() error
	Upgrade(w http.ResponseWriter, r *http.Request) error
	Stats() Stats
	IsAdmin(password string) (bool, error)
}
