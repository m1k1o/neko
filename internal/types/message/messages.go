package message

import (
	"demodesk/neko/internal/types"
)

type Message struct {
	Event string `json:"event"`
}

type Disconnect struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

type SignalProvide struct {
	Event string   `json:"event"`
	ID    string   `json:"id"`
	SDP   string   `json:"sdp"`
	Lite  bool     `json:"lite"`
	ICE   []string `json:"ice"`
}

type SignalAnswer struct {
	Event       string `json:"event"`
	DisplayName string `json:"displayname"`
	SDP         string `json:"sdp"`
}

type MembersList struct {
	Event    string              `json:"event"`
	Memebers []*MembersListEntry `json:"members"`
}

type MembersListEntry struct {
	ID    string `json:"id"`
	Name  string `json:"displayname"`
	Admin bool   `json:"admin"`
}

type Member struct {
	Event string `json:"event"`
	ID    string `json:"id"`
	Name  string `json:"displayname"`
	Admin bool   `json:"admin"`
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
	ScrollLock *bool   `json:"scrollLock,omitempty"`
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

type Admin struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type AdminTarget struct {
	Event  string `json:"event"`
	Target string `json:"target"`
	ID     string `json:"id"`
}

type ScreenResolution struct {
	Event  string `json:"event"`
	ID     string `json:"id,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Rate   int    `json:"rate"`
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
	Event  string `json:"event"`
	URL    string `json:"url"`
}
