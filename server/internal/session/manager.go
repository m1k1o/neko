package session

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/kataras/go-events"
	"github.com/pion/webrtc/v2"

	"n.eko.moe/neko/internal/utils"
)

func New() *SessionManager {
	return &SessionManager{
		host:    "",
		members: make(map[string]*Session),
		emmiter: events.New(),
	}
}

type SessionManager struct {
	host    string
	members map[string]*Session
	emmiter events.EventEmmiter
}

func (m *SessionManager) New(id string, admin bool, socket *websocket.Conn) *Session {
	session := &Session{
		ID:        id,
		Admin:     admin,
		socket:    socket,
		connected: false,
	}

	m.members[id] = session
	m.emmiter.Emit("created", id, session)

	return session
}

func (m *SessionManager) IsHost(id string) bool {
	return m.host == id
}

func (m *SessionManager) HasHost() bool {
	return m.host != ""
}

func (m *SessionManager) SetHost(id string) error {
	_, ok := m.members[id]
	if ok {
		m.host = id
		m.emmiter.Emit("host", id)
		return nil
	}
	return fmt.Errorf("invalid session id %s", id)
}

func (m *SessionManager) GetHost() (*Session, bool) {
	host, ok := m.members[m.host]
	return host, ok
}

func (m *SessionManager) ClearHost() {
	id := m.host
	m.host = ""
	m.emmiter.Emit("host_cleared", id)
}

func (m *SessionManager) Has(id string) bool {
	_, ok := m.members[id]
	return ok
}

func (m *SessionManager) Get(id string) (*Session, bool) {
	session, ok := m.members[id]
	return session, ok
}

func (m *SessionManager) GetConnected() []*Session {
	var sessions []*Session
	for _, sess := range m.members {
		if sess.connected {
			sessions = append(sessions, sess)
		}
	}

	return sessions
}

func (m *SessionManager) Set(id string, session *Session) {
	m.members[id] = session
}

func (m *SessionManager) Destroy(id string) error {
	session, ok := m.members[id]
	if ok {
		err := session.destroy()
		delete(m.members, id)
		m.emmiter.Emit("destroyed", id)
		return err
	}
	return nil
}

func (m *SessionManager) SetSocket(id string, socket *websocket.Conn) (bool, error) {
	session, ok := m.members[id]
	if ok {
		session.socket = socket
		return true, nil
	}

	return false, fmt.Errorf("invalid session id %s", id)
}

func (m *SessionManager) SetPeer(id string, peer *webrtc.PeerConnection) (bool, error) {
	session, ok := m.members[id]
	if ok {
		session.peer = peer
		return true, nil
	}

	return false, fmt.Errorf("invalid session id %s", id)
}

func (m *SessionManager) SetName(id string, name string) (bool, error) {
	session, ok := m.members[id]
	if ok {
		session.Name = name
		session.connected = true
		m.emmiter.Emit("connected", id, session)
		return true, nil
	}

	return false, fmt.Errorf("invalid session id %s", id)
}

func (m *SessionManager) Mute(id string) error {
	session, ok := m.members[id]
	if ok {
		session.Muted = true
	}
	return nil
}

func (m *SessionManager) Unmute(id string) error {
	session, ok := m.members[id]
	if ok {
		session.Muted = false
	}
	return nil
}

func (m *SessionManager) Kick(id string, v interface{}) error {
	session, ok := m.members[id]
	if ok {
		return session.Kick(v)
	}
	return nil
}

func (m *SessionManager) Clear() error {
	return nil
}

func (m *SessionManager) Brodcast(v interface{}, exclude interface{}) error {
	for id, sess := range m.members {
		if !sess.connected {
			continue
		}

		if exclude != nil {
			if in, _ := utils.ArrayIn(id, exclude); in {
				continue
			}
		}

		if err := sess.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (m *SessionManager) OnHost(listener func(id string)) {
	m.emmiter.On("host", func(payload ...interface{}) {
		listener(payload[0].(string))
	})
}

func (m *SessionManager) OnHostCleared(listener func(id string)) {
	m.emmiter.On("host_cleared", func(payload ...interface{}) {
		listener(payload[0].(string))
	})
}

func (m *SessionManager) OnCreated(listener func(id string, session *Session)) {
	m.emmiter.On("created", func(payload ...interface{}) {
		listener(payload[0].(string), payload[1].(*Session))
	})
}

func (m *SessionManager) OnConnected(listener func(id string, session *Session)) {
	m.emmiter.On("connected", func(payload ...interface{}) {
		listener(payload[0].(string), payload[1].(*Session))
	})
}

func (m *SessionManager) OnDestroy(listener func(id string)) {
	m.emmiter.On("destroyed", func(payload ...interface{}) {
		listener(payload[0].(string))
	})
}
