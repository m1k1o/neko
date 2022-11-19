package message

import (
	"m1k1o/neko/internal/types"

	"github.com/pion/webrtc/v3"
)

type Message struct {
	Event string `json:"event"`
}

type SystemInit struct {
	Event           string            `json:"event"`
	ImplicitHosting bool              `json:"implicit_hosting"`
	Locks           map[string]string `json:"locks"`
	FileTransfer    bool              `json:"file_transfer"`
}

type SystemMessage struct {
	Event   string `json:"event"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type SignalProvide struct {
	Event string             `json:"event"`
	ID    string             `json:"id"`
	SDP   string             `json:"sdp"`
	Lite  bool               `json:"lite"`
	ICE   []webrtc.ICEServer `json:"ice"`
}

type SignalOffer struct {
	Event string `json:"event"`
	SDP   string `json:"sdp"`
}

type SignalAnswer struct {
	Event       string `json:"event"`
	DisplayName string `json:"displayname"`
	SDP         string `json:"sdp"`
}

type SignalCandidate struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type MembersList struct {
	Event   string          `json:"event"`
	Members []*types.Member `json:"members"`
}

type Member struct {
	Event string `json:"event"`
	*types.Member
}
type MemberDisconnected struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type Clipboard struct {
	Event string `json:"event"`
	Text  string `json:"text"`
}

type Keyboard struct {
	Event      string  `json:"event"`
	Layout     *string `json:"layout,omitempty"`
	CapsLock   *bool   `json:"capsLock,omitempty"`
	NumLock    *bool   `json:"numLock,omitempty"`
	ScrollLock *bool   `json:"scrollLock,omitempty"` // TODO: ScrollLock is deprecated.
}

type Control struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type ControlTarget struct {
	Event  string `json:"event"`
	ID     string `json:"id"`
	Target string `json:"target"`
}

type ChatReceive struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

type ChatSend struct {
	Event   string `json:"event"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

type EmoteReceive struct {
	Event string `json:"event"`
	Emote string `json:"emote"`
}

type EmoteSend struct {
	Event string `json:"event"`
	ID    string `json:"id"`
	Emote string `json:"emote"`
}

type FileTransferList struct {
	Event string               `json:"event"`
	Cwd   string               `json:"cwd"`
	Files []types.FileListItem `json:"files"`
}

type Admin struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type AdminTarget struct {
	Event  string `json:"event"`
	Target string `json:"target"`
	ID     string `json:"id"`
}

type AdminLock struct {
	Event    string `json:"event"`
	Resource string `json:"resource"`
	ID       string `json:"id"`
}

type ScreenResolution struct {
	Event  string `json:"event"`
	ID     string `json:"id,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Rate   int16  `json:"rate"`
}

type ScreenConfigurations struct {
	Event          string                            `json:"event"`
	Configurations map[int]types.ScreenConfiguration `json:"configurations"`
}

type BroadcastStatus struct {
	Event    string `json:"event"`
	URL      string `json:"url"`
	IsActive bool   `json:"isActive"`
}

type BroadcastCreate struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}
