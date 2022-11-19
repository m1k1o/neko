package types

import (
	"net/http"
	"time"
)

type Stats struct {
	Connections uint32    `json:"connections"`
	Host        string    `json:"host"`
	Members     []*Member `json:"members"`

	Banned map[string]string `json:"banned"` // IP -> session ID (that banned it)
	Locked map[string]string `json:"locked"` // resource name -> session ID (that locked it)

	ServerStartedAt time.Time  `json:"server_started_at"`
	LastAdminLeftAt *time.Time `json:"last_admin_left_at"`
	LastUserLeftAt  *time.Time `json:"last_user_left_at"`

	ControlProtection bool `json:"control_protection"`
	ImplicitControl   bool `json:"implicit_control"`
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
	IsLocked(resource string) bool
	IsAdmin(password string) (bool, error)

	// File Transfer
	CanTransferFiles(password string) (bool, error)
	FileTransferPath(filename string) string
	FileTransferEnabled() bool
}

type FileListItem struct {
	Filename string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
}
