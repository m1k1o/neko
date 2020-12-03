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

// ---
// profile
// ---

func (session *SessionCtx) VerifySecret(secret string) bool {
	return session.profile.Secret == secret
}

func (session *SessionCtx) Name() string {
	return session.profile.Name
}

func (session *SessionCtx) IsAdmin() bool {
	return session.profile.IsAdmin
}

func (session *SessionCtx) CanLogin() bool {
	return session.profile.CanLogin
}

func (session *SessionCtx) CanConnect() bool {
	return session.profile.CanConnect
}

func (session *SessionCtx) CanWatch() bool {
	return session.profile.CanWatch
}

func (session *SessionCtx) CanHost() bool {
	return session.profile.CanHost
}

func (session *SessionCtx) CanAccessClipboard() bool {
	return session.profile.CanAccessClipboard
}

func (session *SessionCtx) SetProfile(profile types.MemberProfile) {
	session.profile = profile
	session.manager.emmiter.Emit("profile_changed", session)
}

// ---
// runtime
// ---

func (session *SessionCtx) IsHost() bool {
	return session.manager.host != nil && session.manager.host.ID() == session.ID()
}

func (session *SessionCtx) IsConnected() bool {
	return session.websocket_connected
}

func (session *SessionCtx) IsReceiving() bool {
	return session.webrtc_connected
}

func (session *SessionCtx) Disconnect(reason string) error {
	if err := session.Send(
		message.SystemDisconnect{
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

// ---
// webscoket
// ---

func (session *SessionCtx) SetWebSocketPeer(websocket_peer types.WebSocketPeer) {
	session.websocket_peer = websocket_peer
}

func (session *SessionCtx) SetWebSocketConnected(connected bool) {
	session.websocket_connected = connected

	if connected {
		session.manager.emmiter.Emit("connected", session)
		return
	}

	session.manager.emmiter.Emit("disconnected", session)
	session.websocket_peer = nil

	// TODO: Refactor. Only if is WebRTC active.
	if err := session.webrtc_peer.Destroy(); err != nil {
		session.logger.Warn().Err(err).Msgf("webrtc destroy has failed")
	}
}

func (session *SessionCtx) Send(v interface{}) error {
	if session.websocket_peer == nil {
		return nil
	}

	return session.websocket_peer.Send(v)
}

// ---
// webrtc
// ---

func (session *SessionCtx) SetWebRTCPeer(webrtc_peer types.WebRTCPeer) {
	session.webrtc_peer = webrtc_peer
}

func (session *SessionCtx) SetWebRTCConnected(connected bool) {
	session.webrtc_connected = connected
	session.manager.emmiter.Emit("state_changed", session)

	if !connected {
		session.webrtc_peer = nil
	}
}

func (session *SessionCtx) SignalAnswer(sdp string) error {
	if session.webrtc_peer == nil {
		return nil
	}

	return session.webrtc_peer.SignalAnswer(sdp)
}
