package session

import (
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

type MemberProfile struct {
	token            string
	name             string
	is_admin         bool
	enabled          bool
	can_control      bool
	can_watch        bool
	clipboard_access bool
}

type SessionCtx struct {
	id             string
	logger         zerolog.Logger
	manager        *SessionManagerCtx
	profile        MemberProfile
	websocket_peer types.WebSocketPeer
	webrtc_peer    types.WebRTCPeer
	connected      bool
}

func (session *SessionCtx) ID() string {
	return session.id
}

func (session *SessionCtx) Name() string {
	return session.profile.name
}

func (session *SessionCtx) Admin() bool {
	return session.profile.is_admin
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.host != nil && session.manager.host.ID() == session.ID()
}

func (session *SessionCtx) Connected() bool {
	return session.connected
}

func (session *SessionCtx) SetName(name string) {
	session.profile.name = name
}

func (session *SessionCtx) SetWebSocketPeer(websocket_peer types.WebSocketPeer) {
	session.websocket_peer = websocket_peer
	session.manager.emmiter.Emit("created", session)
}

func (session *SessionCtx) SetWebRTCPeer(webrtc_peer types.WebRTCPeer) {
	session.webrtc_peer = webrtc_peer
}

func (session *SessionCtx) SetConnected(connected bool) {
	session.connected = connected

	if connected {
		session.manager.emmiter.Emit("connected", session)
	} else {
		session.manager.emmiter.Emit("disconnected", session)
		session.websocket_peer = nil
	
		// TODO: Refactor.
		_ = session.manager.Destroy(session.id)
	}
}

func (session *SessionCtx) Disconnect(reason string) error {
	session.SetConnected(false)

	return session.Send(
		message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: reason,
		})
}

func (session *SessionCtx) Send(v interface{}) error {
	if session.websocket_peer == nil {
		return nil
	}

	return session.websocket_peer.Send(v)
}

func (session *SessionCtx) SignalAnswer(sdp string) error {
	if session.webrtc_peer == nil {
		return nil
	}

	return session.webrtc_peer.SignalAnswer(sdp)
}

func (session *SessionCtx) destroy() error {
	if session.websocket_peer != nil {
		if err := session.websocket_peer.Destroy(); err != nil {
			return err
		}
	}

	if session.webrtc_peer != nil {
		if err := session.webrtc_peer.Destroy(); err != nil {
			return err
		}
	}

	return nil
}
