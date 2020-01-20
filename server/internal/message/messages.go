package message

import "n.eko.moe/neko/internal/session"

type Message struct {
	Event string `json:"event"`
}

type Disconnect struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

type Identity struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type IdentityDetails struct {
	Event    string `json:"event"`
	Username string `json:"username"`
}

type Signal struct {
	Event string `json:"event"`
	SDP   string `json:"sdp"`
}

type MembersList struct {
	Event    string             `json:"event"`
	Memebers []*session.Session `json:"members"`
}

type Member struct {
	Event string `json:"event"`
	*session.Session
}
type MemberDisconnected struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type Control struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type ChatRecieve struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

type ChatSend struct {
	Event   string `json:"event"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

type EmojiRecieve struct {
	Event string `json:"event"`
	Emoji string `json:"emoji"`
}

type EmojiSend struct {
	Event string `json:"event"`
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
}

type Admin struct {
	Event string `json:"event"`
	ID    string `json:"id"`
}

type AdminSubject struct {
	Event   string `json:"event"`
	Subject string `json:"subject"`
	ID      string `json:"id"`
}
