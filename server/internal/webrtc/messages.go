package webrtc

type message struct {
	Event string `json:"event"`
}

type messageIdentityProvide struct {
	message
	ID string `json:"id"`
}

type messageSDP struct {
	message
	SDP string `json:"sdp"`
}
