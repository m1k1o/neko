package session

import (
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
)

// client is expected to reconnect within 5 second
// if some unexpected websocket disconnect happens
const WS_DELAYED_DURATION = 5 * time.Second

type SessionCtx struct {
	id      string
	token   string
	logger  zerolog.Logger
	manager *SessionManagerCtx
	profile types.MemberProfile
	state   types.SessionState

	websocketPeer types.WebSocketPeer
	websocketMu   sync.Mutex

	// websocket delayed set connected events
	wsDelayedMu    sync.Mutex
	wsDelayedTimer *time.Timer

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

	// update webrtc paused state
	if webrtcPeer := session.GetWebRTCPeer(); webrtcPeer != nil {
		webrtcPeer.SetPaused(session.PrivateModeEnabled())
	}
}

func (session *SessionCtx) State() types.SessionState {
	return session.state
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.isHost(session)
}

func (session *SessionCtx) PrivateModeEnabled() bool {
	return session.manager.Settings().PrivateMode && !session.profile.IsAdmin
}

func (session *SessionCtx) SetCursor(cursor types.Cursor) {
	if session.manager.Settings().InactiveCursors && session.profile.SendsInactiveCursor {
		session.manager.SetCursor(cursor, session)
	}
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

func (session *SessionCtx) SetWebSocketConnected(websocketPeer types.WebSocketPeer, connected bool, delayed bool) {
	session.websocketMu.Lock()
	isCurrentPeer := websocketPeer == session.websocketPeer
	session.websocketMu.Unlock()

	if !isCurrentPeer {
		return
	}

	session.logger.Info().
		Bool("connected", connected).
		Bool("delayed", delayed).
		Msg("set websocket connected")

	//
	// ws delayed
	//

	var wsDelayedTimer *time.Timer

	if delayed {
		wsDelayedTimer = time.AfterFunc(WS_DELAYED_DURATION, func() {
			session.SetWebSocketConnected(websocketPeer, connected, false)
		})
	}

	session.wsDelayedMu.Lock()
	if session.wsDelayedTimer != nil {
		session.wsDelayedTimer.Stop()
	}
	session.wsDelayedTimer = wsDelayedTimer
	session.wsDelayedMu.Unlock()

	if delayed {
		return
	}

	//
	// not delayed
	//

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

func (session *SessionCtx) Send(event string, payload any) {
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

	session.logger.Info().
		Bool("connected", connected).
		Msg("set webrtc connected")

	session.state.IsWatching = connected
	session.manager.emmiter.Emit("state_changed", session)

	if connected {
		return
	}

	session.webrtcMu.Lock()
	isCurrentPeer = webrtcPeer == session.webrtcPeer
	if isCurrentPeer {
		session.webrtcPeer = nil
	}
	session.webrtcMu.Unlock()

	if isCurrentPeer {
		session.Send(event.SIGNAL_CLOSE, nil)
	}
}

func (session *SessionCtx) GetWebRTCPeer() types.WebRTCPeer {
	session.webrtcMu.Lock()
	defer session.webrtcMu.Unlock()

	return session.webrtcPeer
}
