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
	Muted     bool   `json:"-"`
	connected bool
	socket    *websocket.Conn
	peer      *webrtc.PeerConnection
	mu        sync.Mutex
}

func (session *Session) RemoteAddr() *string {
	if session.socket != nil {
		address := session.socket.RemoteAddr().String()
		return &address
	}
	return nil
}

// TODO: write to peer data channel
func (session *Session) Write(v interface{}) error {
	session.mu.Lock()
	defer session.mu.Unlock()
	return nil
}

func (session *Session) Kick(v interface{}) error {
	if err := session.Send(v); err != nil {
		return err
	}

	return session.destroy()
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
