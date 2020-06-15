package message

import (
	"n.eko.moe/neko/internal/types"
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
	Event    string          `json:"event"`
	Memebers []*types.Member `json:"members"`
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
	Event  string `json:"event"`
	Layout string `json:"layout"`
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
