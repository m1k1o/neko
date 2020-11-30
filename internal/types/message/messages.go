package message

import (
	"demodesk/neko/internal/types"
)

type Message struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"` // TODO: New.
}

// System
// TODO: New.
type SystemConnect struct {
	Event            string          `json:"event,omitempty"`
	ControlHost      ControlHost     `json:"control_host"`
	ScreenSize       ScreenSize      `json:"screen_size"`
	Members          []MemberData    `json:"members"`
	ScreenSizesList  *[]ScreenSize   `json:"screen_sizes_list,omitempty"`
	BroadcastStatus  BroadcastStatus `json:"members,omitempty"`
}

// TODO: New.
type SystemDisconnect struct {
	Event   string `json:"event,omitempty"`
	Message string `json:"message"`
}

// Signal
type SignalProvide struct {
	Event string   `json:"event,omitempty"`
	ID    string   `json:"id"` // TODO: Remove
	SDP   string   `json:"sdp"`
	Lite  bool     `json:"lite"`
	ICE   []string `json:"ice"`
}

type SignalAnswer struct {
	Event string `json:"event,omitempty"`
	SDP   string `json:"sdp"`
}


// Member
// TODO: New.
type MemberID struct {
	Event string `json:"event,omitempty"`
	ID    string `json:"id"`
}

// TODO: New.
type MemberData struct {
	Event   string `json:"event,omitempty"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

// Control
// TODO: New.
type ControlHost struct {
	Event   string  `json:"event,omitempty"`
	HasHost bool    `json:"has_host"`
	HostID  string  `json:"id,omitempty"`
}

// TODO: New.
type ControlMove struct {
	Event string `json:"event,omitempty"`
	X     uint16 `json:"x"`
	Y     uint16 `json:"y"`
}

// TODO: New.
type ControlScroll struct {
	Event string `json:"event,omitempty"`
	X     int16  `json:"x"`
	Y     int16  `json:"y"`
}

// TODO: New.
type ControlKey struct {
	Event string `json:"event,omitempty"`
	Key   uint64 `json:"key"`
}

// Screen
// TODO: New.
type ScreenSize struct {
	Event  string `json:"event,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Rate   int    `json:"rate"`
}

// Clipboard
// TODO: New.
type ClipboardData struct {
	Event string `json:"event,omitempty"`
	Text  string `json:"text"`
}

// Keyboard
// TODO: New.
type KeyboardModifiers struct {
	Event      string `json:"event,omitempty"`
	CapsLock   bool   `json:"caps_lock"`
	NumLock    bool   `json:"num_lock"`
	ScrollLock bool   `json:"scroll_lock"`	
}

// TODO: New.
type KeyboardLayout struct {
	Event  string `json:"event,omitempty"`
	Layout string `json:"layout"`
}

type BroadcastStatus struct {
	Event    string `json:"event"`
	URL      string `json:"url"`
	IsActive bool   `json:"is_active"`
}

// TODO: Remove.
type Disconnect struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

// TODO: Remove.
type MembersList struct {
	Event    string              `json:"event"`
	Memebers []*MembersListEntry `json:"members"`
}

// TODO: Remove.
type MembersListEntry struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

// TODO: Remove.
type Member struct {
	Event string `json:"event"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

// TODO: Remove.
type MemberDisconnected struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

// TODO: Remove.
type Clipboard struct {
	Event string `json:"event"`
	Text  string `json:"text"`
}

// TODO: Remove.
type Keyboard struct {
	Event      string  `json:"event"`
	Layout     *string `json:"layout,omitempty"`
	CapsLock   *bool   `json:"capsLock,omitempty"`
	NumLock    *bool   `json:"numLock,omitempty"`
	ScrollLock *bool   `json:"scrollLock,omitempty"`
}

// TODO: Remove.
type Control struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

// TODO: Remove.
type ControlTarget struct {
	Event  string `json:"event"`
	ID     string `json:"id"`
	Target string `json:"target"`
}

// TODO: Remove.
type ScreenResolution struct {
	Event  string `json:"event"`
	ID     string `json:"id,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Rate   int    `json:"rate"`
}

// TODO: Remove.
type ScreenConfigurations struct {
	Event          string                            `json:"event"`
	Configurations map[int]types.ScreenConfiguration `json:"configurations"`
}
