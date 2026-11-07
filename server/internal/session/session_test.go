package session

import (
	"testing"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
)

// fakePeer embeds the real interface so we only implement the methods sessions call.
type fakePeer struct {
	types.WebSocketPeer
	destroyed bool
}

func (p *fakePeer) Destroy(string) { p.destroyed = true }

func (p *fakePeer) Send(string, any) {}

func newTestSession(t *testing.T, profile types.MemberProfile) (*SessionManagerCtx, types.Session) {
	t.Helper()

	manager := New(&config.Session{})
	session, _, err := manager.Create("test-session", profile)
	if err != nil {
		t.Fatalf("could not create session: %v", err)
	}

	return manager, session
}

func assertTotals(t *testing.T, manager *SessionManagerCtx, users, admins int) {
	t.Helper()

	stats := manager.Stats()
	if stats.TotalUsers != users || stats.TotalAdmins != admins {
		t.Errorf("got %d users and %d admins, want %d users and %d admins",
			stats.TotalUsers, stats.TotalAdmins, users, admins)
	}
}

func TestStatsCountConnectAndDisconnect(t *testing.T) {
	manager, session := newTestSession(t, types.MemberProfile{})
	peer := &fakePeer{}

	session.ConnectWebSocketPeer(peer)
	assertTotals(t, manager, 1, 0)
	if manager.Stats().LastUserLeftAt != nil {
		t.Error("last user left time is set while a user is connected")
	}

	session.DisconnectWebSocketPeer(peer, false)
	assertTotals(t, manager, 0, 0)
	if manager.Stats().LastUserLeftAt == nil {
		t.Error("last user left time is not set after the last user left")
	}
}

func TestStatsCountReplacedPeerOnce(t *testing.T) {
	tests := []struct {
		name    string
		profile types.MemberProfile
		users   int
		admins  int
	}{
		{"user", types.MemberProfile{}, 1, 0},
		{"admin", types.MemberProfile{IsAdmin: true}, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, session := newTestSession(t, tt.profile)
			first, second := &fakePeer{}, &fakePeer{}

			session.ConnectWebSocketPeer(first)
			session.ConnectWebSocketPeer(second)
			assertTotals(t, manager, tt.users, tt.admins)

			if !first.destroyed {
				t.Error("replaced peer was not destroyed")
			}

			// the replaced peer's connection ending must not count as leaving
			session.DisconnectWebSocketPeer(first, false)
			assertTotals(t, manager, tt.users, tt.admins)

			session.DisconnectWebSocketPeer(second, false)
			assertTotals(t, manager, 0, 0)

			stats := manager.Stats()
			if tt.admins > 0 && stats.LastAdminLeftAt == nil {
				t.Error("last admin left time is not set after the last admin left")
			}
			if tt.users > 0 && stats.LastUserLeftAt == nil {
				t.Error("last user left time is not set after the last user left")
			}
		})
	}
}

func TestStatsFollowAdminProfileChange(t *testing.T) {
	manager, session := newTestSession(t, types.MemberProfile{})
	peer := &fakePeer{}

	session.ConnectWebSocketPeer(peer)
	assertTotals(t, manager, 1, 0)

	// promote
	if err := manager.Update(session.ID(), types.MemberProfile{IsAdmin: true, CanConnect: true, CanLogin: true}); err != nil {
		t.Fatalf("could not update session: %v", err)
	}
	assertTotals(t, manager, 0, 1)
	if manager.Stats().LastUserLeftAt == nil {
		t.Error("last user left time is not set after the only user was promoted")
	}

	// unrelated change must not move the counters
	if err := manager.Update(session.ID(), types.MemberProfile{IsAdmin: true, CanConnect: true, CanLogin: true, Name: "new"}); err != nil {
		t.Fatalf("could not update session: %v", err)
	}
	assertTotals(t, manager, 0, 1)

	// demote
	if err := manager.Update(session.ID(), types.MemberProfile{CanConnect: true, CanLogin: true}); err != nil {
		t.Fatalf("could not update session: %v", err)
	}
	assertTotals(t, manager, 1, 0)

	session.DisconnectWebSocketPeer(peer, false)
	assertTotals(t, manager, 0, 0)
}

func TestStatsIgnoreProfileChangeWhileDisconnected(t *testing.T) {
	manager, session := newTestSession(t, types.MemberProfile{})

	if err := manager.Update(session.ID(), types.MemberProfile{IsAdmin: true}); err != nil {
		t.Fatalf("could not update session: %v", err)
	}
	assertTotals(t, manager, 0, 0)
	if stats := manager.Stats(); stats.LastUserLeftAt != nil || stats.LastAdminLeftAt != nil {
		t.Error("left times are set although nobody was connected")
	}
}
