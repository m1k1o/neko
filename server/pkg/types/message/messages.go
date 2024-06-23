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
	ScreenSize        types.ScreenSize       `json:"screen_size"`
	Sessions          map[string]SessionData `json:"sessions"`
	Settings          types.Settings         `json:"settings"`
	TouchEvents       bool                   `json:"touch_events"`
	ScreencastEnabled bool                   `json:"screencast_enabled"`
	WebRTC            SystemWebRTC           `json:"webrtc"`
}

type SystemAdmin struct {
	ScreenSizesList []types.ScreenSize `json:"screen_sizes_list"`
	BroadcastStatus BroadcastStatus    `json:"broadcast_status"`
}

type SystemLogs = []SystemLog

type SystemLog struct {
	Level   string         `json:"level"`
	Fields  map[string]any `json:"fields"`
	Message string         `json:"message"`
}

type SystemDisconnect struct {
	Message string `json:"message"`
}

type SystemSettingsUpdate struct {
	ID string `json:"id"`
	types.Settings
}

/////////////////////////////
// Signal
/////////////////////////////

type SignalRequest struct {
	Video types.PeerVideoRequest `json:"video"`
	Audio types.PeerAudioRequest `json:"audio"`

	Auto bool `json:"auto"` // TODO: Remove this
}

type SignalProvide struct {
	SDP        string            `json:"sdp"`
	ICEServers []types.ICEServer `json:"iceservers"`

	Video types.PeerVideo `json:"video"`
	Audio types.PeerAudio `json:"audio"`
}

type SignalCandidate struct {
	webrtc.ICECandidateInit
}

type SignalDescription struct {
	SDP string `json:"sdp"`
}

type SignalVideo struct {
	types.PeerVideoRequest
}

type SignalAudio struct {
	types.PeerAudioRequest
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
	ID      string `json:"id"`
	HasHost bool   `json:"has_host"`
	HostID  string `json:"host_id,omitempty"`
}

type ControlScroll struct {
	// TOOD: remove this once the client is fixed
	X int `json:"x"`
	Y int `json:"y"`

	DeltaX     int  `json:"delta_x"`
	DeltaY     int  `json:"delta_y"`
	ControlKey bool `json:"control_key"`
}

type ControlPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ControlButton struct {
	*ControlPos
	Code uint32 `json:"code"`
}

type ControlKey struct {
	*ControlPos
	Keysym uint32 `json:"keysym"`
}

type ControlTouch struct {
	TouchId uint32 `json:"touch_id"`
	*ControlPos
	Pressure uint8 `json:"pressure"`
}

/////////////////////////////
// Screen
/////////////////////////////

type ScreenSize struct {
	types.ScreenSize
}

type ScreenSizeUpdate struct {
	ID string `json:"id"`
	types.ScreenSize
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
	types.KeyboardMap
}

type KeyboardModifiers struct {
	types.KeyboardModifiers
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
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Subject  string `json:"subject"`
	Body     any    `json:"body"`
}

type SendBroadcast struct {
	Sender  string `json:"sender"`
	Subject string `json:"subject"`
	Body    any    `json:"body"`
}
