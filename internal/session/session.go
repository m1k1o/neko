package session

import (
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

type SessionCtx struct {
	id                  string
	logger              zerolog.Logger
	manager             *SessionManagerCtx
	profile             types.MemberProfile
	websocket_peer      types.WebSocketPeer
	websocket_connected bool
	webrtc_peer         types.WebRTCPeer
	webrtc_connected    bool
}

func (session *SessionCtx) ID() string {
	return session.id
}

func (session *SessionCtx) Name() string {
	return session.profile.Name
}

func (session *SessionCtx) Admin() bool {
	return session.profile.IsAdmin
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.host != nil && session.manager.host.ID() == session.ID()
}

func (session *SessionCtx) VerifySecret(secret string) bool {
	return session.profile.Secret == secret
}

func (session *SessionCtx) Connected() bool {
	return session.websocket_connected && session.webrtc_connected
}

func (session *SessionCtx) SetWebSocketPeer(websocket_peer types.WebSocketPeer) {
	session.websocket_peer = websocket_peer
}

func (session *SessionCtx) SetWebSocketConnected(connected bool) {
	if connected {
		session.websocket_connected = true
		session.manager.emmiter.Emit("websocket_connected", session)
	} else {
		session.websocket_connected = false

		// TODO: Refactor.
		//session.manager.emmiter.Emit("websocket_disconnected", session)
		session.manager.emmiter.Emit("disconnected", session)

		session.websocket_peer = nil
	}
}

func (session *SessionCtx) SetWebRTCPeer(webrtc_peer types.WebRTCPeer) {
	session.webrtc_peer = webrtc_peer
}

func (session *SessionCtx) SetWebRTCConnected(connected bool) {
	if connected {
		session.webrtc_connected = true

		// TODO: Refactor.
		//session.manager.emmiter.Emit("webrtc_connected", session)
		session.manager.emmiter.Emit("connected", session)
	} else {
		session.webrtc_connected = false
		session.manager.emmiter.Emit("webrtc_disconnected", session)

		session.webrtc_peer = nil
	}
}

func (session *SessionCtx) Disconnect(reason string) error {
	if err := session.Send(
		message.Disconnect{
			Event:   event.SYSTEM_DISCONNECT,
			Message: reason,
		}); err != nil {
		return err
	}

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
