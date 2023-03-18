package session

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"

	"github.com/rs/zerolog"
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

func (session *Session) SetConnected(connected bool) error {
	session.connected = connected
	if connected {
		session.manager.eventsChannel <- types.SessionEvent{
			Type:    types.SESSION_CONNECTED,
			Id:      session.id,
			Session: session,
		}
	}
	return nil
}

func (session *Session) Kick(reason string) error {
	if session.socket == nil {
		return nil
	}
	if err := session.socket.Send(&message.SystemMessage{
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

func (session *Session) SignalLocalOffer(sdp string) error {
	if session.peer == nil {
		return nil
	}
	session.logger.Info().Msg("signal update - LocalOffer")
	return session.socket.Send(&message.SignalOffer{
		Event: event.SIGNAL_OFFER,
		SDP:   sdp,
	})
}

func (session *Session) SignalLocalAnswer(sdp string) error {
	if session.peer == nil {
		return nil
	}

	session.logger.Info().Msg("signal update - LocalAnswer")
	return session.socket.Send(&message.SignalAnswer{
		Event: event.SIGNAL_ANSWER,
		SDP:   sdp,
	})
}

func (session *Session) SignalLocalCandidate(data string) error {
	if session.socket == nil {
		return nil
	}
	session.logger.Info().Msg("signal update - LocalCandidate")
	return session.socket.Send(&message.SignalCandidate{
		Event: event.SIGNAL_CANDIDATE,
		Data:  data,
	})
}

func (session *Session) SignalRemoteOffer(sdp string) error {
	if session.peer == nil {
		return nil
	}
	if err := session.peer.SetOffer(sdp); err != nil {
		return err
	}
	sdp, err := session.peer.CreateAnswer()
	if err != nil {
		return err
	}
	session.logger.Info().Msg("signal update - RemoteOffer")
	return session.SignalLocalAnswer(sdp)
}

func (session *Session) SignalRemoteAnswer(sdp string) error {
	if session.peer == nil {
		return nil
	}
	session.logger.Info().Msg("signal update - RemoteAnswer")
	return session.peer.SetAnswer(sdp)
}

func (session *Session) SignalRemoteCandidate(data string) error {
	if session.socket == nil {
		return nil
	}
	session.logger.Info().Msg("signal update - RemoteCandidate")
	return session.peer.SetCandidate(data)
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
