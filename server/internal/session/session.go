package session

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
)

type Session struct {
	ID        string `json:"id"`
	Name      string `json:"username"`
	Admin     bool   `json:"admin"`
	connected bool
	socket    *websocket.Conn
	peer      *webrtc.PeerConnection
	mu        sync.Mutex
}

// TODO: write to peer data channel
func (session *Session) Write(v interface{}) error {
	session.mu.Lock()
	defer session.mu.Unlock()
	return nil
}

func (session *Session) Send(v interface{}) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.socket != nil {
		return session.socket.WriteJSON(v)
	}

	return nil
}

func (session *Session) destroy() error {
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
