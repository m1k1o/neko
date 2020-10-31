package session

import (
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

type Session struct {
	logger    zerolog.Logger
	id        string
	name      string
	admin     bool
	muted     bool
	connected bool
	manager   *SessionManager
	socket    types.WebSocket
	peer      types.Peer
}

func (session *Session) ID() string {
	return session.id
}

func (session *Session) Name() string {
	return session.name
}

func (session *Session) Admin() bool {
	return session.admin
}

func (session *Session) Muted() bool {
	return session.muted
}

func (session *Session) IsHost() bool {
	return session.manager.host == session.id
}

func (session *Session) Connected() bool {
	return session.connected
}

func (session *Session) Address() string {
	if session.socket == nil {
		return ""
	}
	return session.socket.Address()
}

func (session *Session) Member() *types.Member {
	return &types.Member{
		ID:    session.id,
		Name:  session.name,
		Admin: session.admin,
		Muted: session.muted,
	}
}

func (session *Session) SetMuted(muted bool) {
	session.muted = muted
}

func (session *Session) SetName(name string) error {
	session.name = name
	return nil
}

func (session *Session) SetSocket(socket types.WebSocket) error {
	session.socket = socket
	return nil
}

func (session *Session) SetPeer(peer types.Peer) error {
	session.peer = peer
	return nil
}

func (session *Session) SetConnected(connected bool) {
	session.connected = connected
	if connected {
		session.manager.emmiter.Emit("connected", session.id, session)
	}
}

func (session *Session) Disconnect(reason string) error {
	if session.socket == nil {
		return nil
	}

	if err := session.socket.Send(&message.Disconnect{
		Event:   event.SYSTEM_DISCONNECT,
		Message: reason,
	}); err != nil {
		return err
	}

	return session.manager.Destroy(session.id)
}

func (session *Session) Send(v interface{}) error {
	if session.socket == nil {
		return nil
	}
	return session.socket.Send(v)
}

func (session *Session) SignalAnswer(sdp string) error {
	if session.peer == nil {
		return nil
	}
	return session.peer.SignalAnswer(sdp)
}

func (session *Session) destroy() error {
	if session.socket != nil {
		if err := session.socket.Destroy(); err != nil {
			return err
		}
	}

	if session.peer != nil {
		if err := session.peer.Destroy(); err != nil {
			return err
		}
	}

	return nil
}
