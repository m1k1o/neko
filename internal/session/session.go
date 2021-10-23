package session

import (
	"sync"

	"github.com/rs/zerolog"

	"demodesk/neko/internal/types"
)

type SessionCtx struct {
	id      string
	token   string
	logger  zerolog.Logger
	manager *SessionManagerCtx
	profile types.MemberProfile
	state   types.SessionState

	positionX  int
	positionY  int
	positionMu sync.Mutex

	websocketPeer types.WebSocketPeer
	websocketMu   sync.Mutex

	webrtcPeer types.WebRTCPeer
	webrtcMu   sync.Mutex
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
		session.GetWebRTCPeer().Destroy()
	}

	if (!session.profile.CanConnect || !session.profile.CanLogin) && session.state.IsConnected {
		session.GetWebSocketPeer().Destroy("profile changed")
	}
}

func (session *SessionCtx) State() types.SessionState {
	return session.state
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.GetHost() == session
}

// ---
// cursor position
// ---

func (session *SessionCtx) SetPosition(x, y int) {
	session.positionMu.Lock()
	defer session.positionMu.Unlock()

	session.positionX = x
	session.positionY = y
}

// ---
// websocket
// ---

func (session *SessionCtx) SetWebSocketPeer(websocketPeer types.WebSocketPeer) {
	session.websocketMu.Lock()
	session.websocketPeer, websocketPeer = websocketPeer, session.websocketPeer
	session.websocketMu.Unlock()

	if websocketPeer != nil && websocketPeer != session.websocketPeer {
		websocketPeer.Destroy("connection replaced")
	}
}

func (session *SessionCtx) SetWebSocketConnected(websocketPeer types.WebSocketPeer, connected bool) {
	session.websocketMu.Lock()
	isCurrentPeer := websocketPeer == session.websocketPeer
	session.websocketMu.Unlock()

	if !isCurrentPeer {
		return
	}

	session.state.IsConnected = connected

	if connected {
		session.manager.emmiter.Emit("connected", session)
		return
	}

	session.manager.emmiter.Emit("disconnected", session)

	session.websocketMu.Lock()
	if websocketPeer == session.websocketPeer {
		session.websocketPeer = nil
	}
	session.websocketMu.Unlock()
}

func (session *SessionCtx) GetWebSocketPeer() types.WebSocketPeer {
	session.websocketMu.Lock()
	defer session.websocketMu.Unlock()

	return session.websocketPeer
}

func (session *SessionCtx) Send(event string, payload interface{}) {
	peer := session.GetWebSocketPeer()
	if peer != nil {
		peer.Send(event, payload)
	}
}

// ---
// webrtc
// ---

func (session *SessionCtx) SetWebRTCPeer(webrtcPeer types.WebRTCPeer) {
	session.webrtcMu.Lock()
	session.webrtcPeer, webrtcPeer = webrtcPeer, session.webrtcPeer
	session.webrtcMu.Unlock()

	if webrtcPeer != nil && webrtcPeer != session.webrtcPeer {
		webrtcPeer.Destroy()
	}
}

func (session *SessionCtx) SetWebRTCConnected(webrtcPeer types.WebRTCPeer, connected bool) {
	session.webrtcMu.Lock()
	isCurrentPeer := webrtcPeer == session.webrtcPeer
	session.webrtcMu.Unlock()

	if !isCurrentPeer {
		return
	}

	session.state.IsWatching = connected
	session.manager.emmiter.Emit("state_changed", session)

	if connected {
		return
	}

	session.webrtcMu.Lock()
	if webrtcPeer == session.webrtcPeer {
		session.webrtcPeer = nil
	}
	session.webrtcMu.Unlock()
}

func (session *SessionCtx) GetWebRTCPeer() types.WebRTCPeer {
	session.webrtcMu.Lock()
	defer session.webrtcMu.Unlock()

	return session.webrtcPeer
}
