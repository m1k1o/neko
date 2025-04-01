package types

import "time"

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

type Member struct {
	ID    string `json:"id"`
	Name  string `json:"displayname"`
	Admin bool   `json:"admin"`
	Muted bool   `json:"muted"`
}

type FileListItem struct {
	Filename string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
}

type ScreenConfiguration struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Rates  map[int]int16 `json:"rates"`
}
