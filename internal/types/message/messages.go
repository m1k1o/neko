package message

type Message struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"` // TODO: New.
}

/////////////////////////////
// System
/////////////////////////////

type SystemInit struct {
	Event           string           `json:"event,omitempty"`
	MemberId        string           `json:"member_id"`
	ControlHost     ControlHost      `json:"control_host"`
	ScreenSize      ScreenSize       `json:"screen_size"`
	Members         []MemberData     `json:"members"`
}

type SystemAdmin struct {
	Event           string           `json:"event,omitempty"`
	ScreenSizesList []ScreenSize     `json:"screen_sizes_list"`
	BroadcastStatus BroadcastStatus  `json:"broadcast_status"`
}

type SystemDisconnect struct {
	Event   string `json:"event,omitempty"`
	Message string `json:"message"`
}

/////////////////////////////
// Signal
/////////////////////////////

type SignalProvide struct {
	Event string   `json:"event,omitempty"`
	SDP   string   `json:"sdp"`
	Lite  bool     `json:"lite"`
	ICE   []string `json:"ice"`
}

type SignalAnswer struct {
	Event string `json:"event,omitempty"`
	SDP   string `json:"sdp"`
}

/////////////////////////////
// Member
/////////////////////////////

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

/////////////////////////////
// Control
/////////////////////////////

type ControlHost struct {
	Event   string  `json:"event,omitempty"`
	HasHost bool    `json:"has_host"`
	HostID  string  `json:"host_id,omitempty"`
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

/////////////////////////////
// Screen
/////////////////////////////

type ScreenSize struct {
	Event  string `json:"event,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Rate   int    `json:"rate"`
}

/////////////////////////////
// Clipboard
/////////////////////////////

type ClipboardData struct {
	Event string `json:"event,omitempty"`
	Text  string `json:"text"`
}

/////////////////////////////
// Keyboard
/////////////////////////////

type KeyboardModifiers struct {
	Event      string `json:"event,omitempty"`
	CapsLock   *bool  `json:"caps_lock"`
	NumLock    *bool  `json:"num_lock"`
	ScrollLock *bool  `json:"scroll_lock"`	
}

type KeyboardLayout struct {
	Event  string `json:"event,omitempty"`
	Layout string `json:"layout"`
}

/////////////////////////////
// Broadcast
/////////////////////////////

type BroadcastStatus struct {
	Event    string `json:"event,omitempty"`
	IsActive bool   `json:"is_active"`
	URL      string `json:"url,omitempty"`
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
