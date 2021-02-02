package session

import (
	"github.com/rs/zerolog"
	"github.com/pion/webrtc/v3"

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

func (session *SessionCtx) GetProfile() types.MemberProfile {
	profile := session.profile
	profile.Secret = ""
	return profile
}

func (session *SessionCtx) profileChanged() {
	if !session.CanHost() && session.IsHost() {
		session.manager.ClearHost()

		session.manager.Broadcast(
			message.ControlHost{
				Event:   event.CONTROL_HOST,
				HasHost: false,
			}, nil)
	}

	if !session.CanWatch() && session.IsWatching() {
		if err := session.webrtc_peer.Destroy(); err != nil {
			session.logger.Warn().Err(err).Msgf("webrtc destroy has failed")
		}
	}

	if (!session.CanConnect() || !session.CanLogin()) && session.IsConnected() {
		if err := session.Disconnect("profile changed"); err != nil {
			session.logger.Warn().Err(err).Msgf("websocket destroy has failed")
		}
	}
}

// ---
// state
// ---

func (session *SessionCtx) IsHost() bool {
	return session.manager.host != nil && session.manager.host.ID() == session.ID()
}

func (session *SessionCtx) IsConnected() bool {
	return session.websocket_connected
}

func (session *SessionCtx) IsWatching() bool {
	return session.webrtc_connected
}

func (session *SessionCtx) GetState() types.MemberState {
	// TODO: Save state in member struct.
	return types.MemberState{
		IsConnected: session.IsConnected(),
		IsWatching:  session.IsWatching(),
	}
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

	if session.webrtc_peer != nil {
		if err := session.webrtc_peer.Destroy(); err != nil {
			session.logger.Warn().Err(err).Msgf("webrtc destroy has failed")
		}
	}
}

func (session *SessionCtx) Send(v interface{}) error {
	if session.websocket_peer == nil {
		return nil
	}

	return session.websocket_peer.Send(v)
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

func (session *SessionCtx) SignalCandidate(candidate webrtc.ICECandidateInit) error {
	if session.webrtc_peer == nil {
		return nil
	}

	return session.webrtc_peer.SignalCandidate(candidate)
}
