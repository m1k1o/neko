package session

import (
	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
)

type SessionCtx struct {
	id            string
	token         string
	logger        zerolog.Logger
	manager       *SessionManagerCtx
	profile       types.MemberProfile
	state         types.SessionState
	websocketPeer types.WebSocketPeer
	webrtcPeer    types.WebRTCPeer
}

func (session *SessionCtx) ID() string {
	return session.id
}

func (session *SessionCtx) Profile() types.MemberProfile {
	return session.profile
}

func (session *SessionCtx) profileChanged() {
	if !session.profile.CanHost && session.IsHost() {
		session.manager.ClearHost()
	}

	if (!session.profile.CanConnect || !session.profile.CanLogin || !session.profile.CanWatch) && session.state.IsWatching {
		session.webrtcPeer.Destroy()
	}

	if (!session.profile.CanConnect || !session.profile.CanLogin) && session.state.IsConnected {
		session.websocketPeer.Destroy("profile changed")
	}
}

func (session *SessionCtx) State() types.SessionState {
	return session.state
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.host != nil && session.manager.host == session
}

// ---
// webscoket
// ---

func (session *SessionCtx) SetWebSocketPeer(websocketPeer types.WebSocketPeer) {
	if session.websocketPeer != nil {
		session.websocketPeer.Destroy("connection replaced")
	}

	session.websocketPeer = websocketPeer
}

func (session *SessionCtx) SetWebSocketConnected(websocketPeer types.WebSocketPeer, connected bool) {
	if websocketPeer != session.websocketPeer {
		return
	}

	session.state.IsConnected = connected

	if connected {
		session.manager.emmiter.Emit("connected", session)
		return
	}

	session.manager.emmiter.Emit("disconnected", session)
	session.websocketPeer = nil
}

func (session *SessionCtx) GetWebSocketPeer() types.WebSocketPeer {
	return session.websocketPeer
}

func (session *SessionCtx) Send(event string, payload interface{}) {
	if session.websocketPeer == nil {
		return
	}

	session.websocketPeer.Send(event, payload)
}

// ---
// webrtc
// ---

func (session *SessionCtx) SetWebRTCPeer(webrtcPeer types.WebRTCPeer) {
	if session.webrtcPeer != nil {
		session.webrtcPeer.Destroy()
	}

	session.webrtcPeer = webrtcPeer
}

func (session *SessionCtx) SetWebRTCConnected(webrtcPeer types.WebRTCPeer, connected bool) {
	if webrtcPeer != session.webrtcPeer {
		return
	}

	session.state.IsWatching = connected
	session.manager.emmiter.Emit("state_changed", session)

	if !connected {
		session.webrtcPeer = nil
	}
}

func (session *SessionCtx) GetWebRTCPeer() types.WebRTCPeer {
	return session.webrtcPeer
}
