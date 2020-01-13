package webrtc

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
)

type session struct {
	id     string
	socket *websocket.Conn
	peer   *webrtc.PeerConnection
	mu     sync.Mutex
}

func (session *session) send(v interface{}) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.socket != nil {
		return session.socket.WriteJSON(v)
	}

	return nil
}

func (session *session) destroy() error {
	if session.peer != nil && session.peer.ConnectionState() == webrtc.PeerConnectionStateConnected {
		if err := session.peer.Close(); err != nil {
			return err
		}
	}

	if session.socket != nil {
		if err := session.socket.Close(); err != nil {
			return err
		}
	}
	return nil
}
