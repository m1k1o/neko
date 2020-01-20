package message

import "n.eko.moe/neko/internal/session"

type Message struct {
	Event string `json:"event"`
}

type Identity struct {
	Message
	ID string `json:"id"`
}

type IdentityDetails struct {
	Message
	Username string `json:"username"`
}

type Signal struct {
	Message
	SDP string `json:"sdp"`
}

type Members struct {
	Message
	Memebers []*session.Session `json:"members"`
}

type Member struct {
	Message
	*session.Session
}
type MemberDisconnected struct {
	Message
	ID string `json:"id"`
}

type Control struct {
	Message
	ID string `json:"id"`
}

type Chat struct {
	Message
	ID      string `json:"id"`
	Content string `json:"content"`
}
