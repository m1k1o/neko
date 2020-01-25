package session

import (
	"sync"

	"github.com/rs/zerolog"
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/event"
	"n.eko.moe/neko/internal/types/message"
)

type Session struct {
	logger    zerolog.Logger
	id        string
	name      string
	admin     bool
	muted     bool
	connected bool
	manager   *SessionManager
	socket    types.WebScoket
	peer      types.Peer
	mu        sync.Mutex
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

func (session *Session) Connected() bool {
	return session.connected
}

func (session *Session) Address() *string {
	if session.socket == nil {
		return nil
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
	session.connected = true
	session.manager.emmiter.Emit("connected", session.id, session)
	return nil
}

func (session *Session) SetSocket(socket types.WebScoket) error {
	session.socket = socket
	return nil
}

func (session *Session) SetPeer(peer types.Peer) error {
	session.peer = peer
	return nil
}

func (session *Session) Kick(reason string) error {
	if session.socket == nil {
		return nil
	}
	if err := session.socket.Send(&message.Disconnect{
		Event:   event.SYSTEM_DISCONNECT,
		Message: reason,
	}); err != nil {
		return err
	}

	return session.destroy()
}

func (session *Session) Send(v interface{}) error {
	if session.socket == nil {
		return nil
	}
	return session.socket.Send(v)
}

func (session *Session) Write(v interface{}) error {
	if session.socket == nil {
		return nil
	}
	return session.socket.Send(v)
}

func (session *Session) WriteVideoSample(sample types.Sample) error {
	if session.peer == nil || !session.connected {
		return nil
	}
	return session.peer.WriteVideoSample(sample)
}

func (session *Session) WriteAudioSample(sample types.Sample) error {
	if session.peer == nil || !session.connected {
		return nil
	}
	return session.peer.WriteAudioSample(sample)
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
