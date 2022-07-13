package message

import (
	"github.com/pion/webrtc/v3"

	"github.com/demodesk/neko/pkg/types"
)

/////////////////////////////
// System
/////////////////////////////

type SystemWebRTC struct {
	Videos []string `json:"videos"`
}

type SystemInit struct {
	SessionId         string                 `json:"session_id"`
	ControlHost       ControlHost            `json:"control_host"`
	ScreenSize        ScreenSize             `json:"screen_size"`
	Sessions          map[string]SessionData `json:"sessions"`
	Settings          types.Settings         `json:"settings"`
	ScreencastEnabled bool                   `json:"screencast_enabled"`
	WebRTC            SystemWebRTC           `json:"webrtc"`
}

type SystemAdmin struct {
	ScreenSizesList []ScreenSize    `json:"screen_sizes_list"`
	BroadcastStatus BroadcastStatus `json:"broadcast_status"`
}

type SystemLogs = []SystemLog

type SystemLog struct {
	Level   string                 `json:"level"`
	Fields  map[string]interface{} `json:"fields"`
	Message string                 `json:"message"`
}

type SystemDisconnect struct {
	Message string `json:"message"`
}

/////////////////////////////
// Signal
/////////////////////////////

type SignalProvide struct {
	SDP        string            `json:"sdp"`
	ICEServers []types.ICEServer `json:"iceservers"`
	Video      string            `json:"video"`
}

type SignalCandidate struct {
	webrtc.ICECandidateInit
}

type SignalDescription struct {
	SDP string `json:"sdp"`
}

type SignalVideo struct {
	Video string `json:"video"`
}

/////////////////////////////
// Session
/////////////////////////////

type SessionID struct {
	ID string `json:"id"`
}

type MemberProfile struct {
	ID string `json:"id"`
	types.MemberProfile
}

type SessionState struct {
	ID string `json:"id"`
	types.SessionState
}

type SessionData struct {
	ID      string              `json:"id"`
	Profile types.MemberProfile `json:"profile"`
	State   types.SessionState  `json:"state"`
}

type SessionCursors struct {
	ID      string         `json:"id"`
	Cursors []types.Cursor `json:"cursors"`
}

/////////////////////////////
// Control
/////////////////////////////

type ControlHost struct {
	HasHost bool   `json:"has_host"`
	HostID  string `json:"host_id,omitempty"`
}

// TODO: New.
type ControlMove struct {
	X uint16 `json:"x"`
	Y uint16 `json:"y"`
}

// TODO: New.
type ControlScroll struct {
	X int16 `json:"x"`
	Y int16 `json:"y"`
}

type ControlKey struct {
	Keysym uint32 `json:"keysym"`
}

/////////////////////////////
// Screen
/////////////////////////////

type ScreenSize struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rate   int16 `json:"rate"`
}

/////////////////////////////
// Clipboard
/////////////////////////////

type ClipboardData struct {
	Text string `json:"text"`
}

/////////////////////////////
// Keyboard
/////////////////////////////

type KeyboardMap struct {
	Layout  string `json:"layout"`
	Variant string `json:"variant"`
}

type KeyboardModifiers struct {
	CapsLock *bool `json:"capslock"`
	NumLock  *bool `json:"numlock"`
}

/////////////////////////////
// Broadcast
/////////////////////////////

type BroadcastStatus struct {
	IsActive bool   `json:"is_active"`
	URL      string `json:"url,omitempty"`
}

/////////////////////////////
// Send (opaque comunication channel)
/////////////////////////////

type SendUnicast struct {
	Sender   string      `json:"sender"`
	Receiver string      `json:"receiver"`
	Subject  string      `json:"subject"`
	Body     interface{} `json:"body"`
}

type SendBroadcast struct {
	Sender  string      `json:"sender"`
	Subject string      `json:"subject"`
	Body    interface{} `json:"body"`
}
