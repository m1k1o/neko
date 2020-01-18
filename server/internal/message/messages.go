package message

type Message struct {
	Event string `json:"event"`
}

type IdentityProvide struct {
	Message
	ID string `json:"id"`
}

type SDP struct {
	Message
	SDP string `json:"sdp"`
}
