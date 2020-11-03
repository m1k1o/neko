package session

import (
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

type SessionCtx struct {
	logger    zerolog.Logger
	id        string
	name      string
	admin     bool
	muted     bool
	connected bool
	manager   *SessionManagerCtx
	socket    types.WebSocket
	peer      types.Peer
}

func (session *SessionCtx) ID() string {
	return session.id
}

func (session *SessionCtx) Name() string {
	return session.name
}

func (session *SessionCtx) Admin() bool {
	return session.admin
}

func (session *SessionCtx) Muted() bool {
	return session.muted
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.host != nil && session.manager.host.ID() == session.ID()
}

func (session *SessionCtx) Connected() bool {
	return session.connected
}

func (session *SessionCtx) Address() string {
	if session.socket == nil {
		return ""
	}

	return session.socket.Address()
}

func (session *SessionCtx) SetMuted(muted bool) {
	session.muted = muted
}

func (session *SessionCtx) SetName(name string) {
	session.name = name
}

func (session *SessionCtx) SetSocket(socket types.WebSocket) {
	session.socket = socket
	session.manager.emmiter.Emit("created", session)
}

func (session *SessionCtx) SetPeer(peer types.Peer) {
	session.peer = peer
}

func (session *SessionCtx) SetConnected(connected bool) {
	session.connected = connected

	if connected {
		session.manager.emmiter.Emit("connected", session)
	} else {
		session.manager.emmiter.Emit("disconnected", session)
		session.socket = nil
	
		// TODO: Refactor.
		_ = session.manager.Destroy(session.id)
	}
}

func (session *SessionCtx) Disconnect(reason string) error {
	if session.socket == nil {
		return nil
	}

	// TODO: Refcator
	if err := session.socket.Send(&message.Disconnect{
		Event:   event.SYSTEM_DISCONNECT,
		Message: reason,
	}); err != nil {
		return err
	}

	session.SetConnected(false)
	return nil
}

func (session *SessionCtx) Send(v interface{}) error {
	if session.socket == nil {
		return nil
	}

	return session.socket.Send(v)
}

func (session *SessionCtx) SignalAnswer(sdp string) error {
	if session.peer == nil {
		return nil
	}

	return session.peer.SignalAnswer(sdp)
}

func (session *SessionCtx) destroy() error {
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
