package websocket

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/internal/websocket/handler"
	"github.com/m1k1o/neko/server/pkg/types"
)

const (
	testTimeout = 5 * time.Second
	readyEvent  = "test/ready"
)

// The fakes below embed the real interfaces so we only implement the methods
// the connection lifecycle actually calls.
type fakeSessions struct {
	types.SessionManager
	session types.Session
}

func (s *fakeSessions) Authenticate(*http.Request) (types.Session, error) { return s.session, nil }

type fakeSession struct {
	types.Session
	peer         types.WebSocketPeer
	disconnected chan struct{}
}

func (s *fakeSession) ID() string { return "test-session" }

func (s *fakeSession) Profile() types.MemberProfile {
	return types.MemberProfile{CanConnect: true}
}

func (s *fakeSession) State() types.SessionState { return types.SessionState{} }

func (s *fakeSession) ConnectWebSocketPeer(peer types.WebSocketPeer) { s.peer = peer }

func (s *fakeSession) DisconnectWebSocketPeer(types.WebSocketPeer, bool) { close(s.disconnected) }

// closeTrackingListener lets a test observe the server side of a hijacked connection.
type closeTrackingListener struct {
	net.Listener
	closed chan struct{}
}

func (l *closeTrackingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &closeTrackingConn{Conn: conn, closed: l.closed}, nil
}

type closeTrackingConn struct {
	net.Conn
	once   sync.Once
	closed chan struct{}
}

func (c *closeTrackingConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return c.Conn.Close()
}

// logBuffer collects log output written from the connection goroutines.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *logBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// testConnection is a real websocket connection handled by the manager.
type testConnection struct {
	manager *WebSocketManagerCtx
	session *fakeSession
	client  *websocket.Conn
	logs    *logBuffer
	// closed when the server closes its side of the connection
	serverConnClosed chan struct{}
}

// connectClient runs a real websocket connection through Upgrade and returns once the
// server is handling its messages.
func connectClient(t *testing.T) *testConnection {
	t.Helper()

	logs := &logBuffer{}
	session := &fakeSession{disconnected: make(chan struct{})}
	sessions := &fakeSessions{session: session}
	manager := &WebSocketManagerCtx{
		logger:   zerolog.New(logs),
		shutdown: make(chan struct{}),
		sessions: sessions,
		handler:  handler.New(sessions, nil, nil, nil),
	}

	ready := make(chan struct{})
	manager.AddHandler(func(_ types.Session, msg types.WebSocketMessage) bool {
		if msg.Event == readyEvent {
			close(ready)
		}
		return true
	})

	upgrade := manager.Upgrade(func(*http.Request) bool { return true })
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := upgrade(w, r); err != nil {
			t.Errorf("upgrade failed: %v", err)
		}
	}))

	serverConnClosed := make(chan struct{})
	server.Listener = &closeTrackingListener{Listener: server.Listener, closed: serverConnClosed}
	server.Start()
	t.Cleanup(server.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	if err := client.WriteJSON(types.WebSocketMessage{Event: readyEvent}); err != nil {
		t.Fatalf("sending message failed: %v", err)
	}
	waitFor(t, ready, "server to handle a message")

	return &testConnection{
		manager:          manager,
		session:          session,
		client:           client,
		logs:             logs,
		serverConnClosed: serverConnClosed,
	}
}

func waitFor(t *testing.T, done <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestShutdownWithConnectedClient(t *testing.T) {
	conn := connectClient(t)

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := conn.manager.Shutdown(); err != nil {
			t.Errorf("shutdown failed: %v", err)
		}
	}()

	waitFor(t, done, "shutdown to finish")
	waitFor(t, conn.session.disconnected, "peer to disconnect")
	waitFor(t, conn.serverConnClosed, "server to close the connection")
}

func closeClient(t *testing.T, conn *testConnection) {
	t.Helper()

	err := conn.client.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		t.Fatalf("sending close frame failed: %v", err)
	}

	waitFor(t, conn.session.disconnected, "peer to disconnect")
	waitFor(t, conn.serverConnClosed, "server to close the connection")
}

func TestClientCloseReleasesServerConnection(t *testing.T) {
	conn := connectClient(t)
	closeClient(t, conn)

	// the read loop has finished, so there is nothing left to wait for
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn.manager.wg.Wait()
	}()
	waitFor(t, done, "read loop to finish")
}

// A peer is replaced after its connection ended when the client reconnects
// within the delayed disconnect window.
func TestDestroyPeerAfterConnectionEnded(t *testing.T) {
	conn := connectClient(t)
	closeClient(t, conn)

	conn.session.peer.Destroy("connection replaced")

	if logs := conn.logs.String(); strings.Contains(logs, `"level":"error"`) {
		t.Errorf("destroying the peer logged an error: %s", logs)
	}
}
