package webrtc

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
)

type fakeSession struct {
	types.Session
	id string
}

func (s *fakeSession) ID() string { return s.id }

func TestTURNCredentials(t *testing.T) {
	expires := time.Unix(1700086400, 0)

	// expected credentials computed independently:
	// base64(hmac_sha1(secret, username))
	tests := []struct {
		name               string
		user               string
		expectedUsername   string
		expectedCredential string
	}{
		{"with user", "alice", "1700086400:alice", "SXua5ne/+mDhiHTp0pQJzRO4ESg="},
		{"without user", "", "1700086400", "K+RlgFj3QLvB7WAVHfjfgMgyBeA="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, credential := turnCredentials("north", tt.user, expires)
			if username != tt.expectedUsername {
				t.Errorf("username = %q, want %q", username, tt.expectedUsername)
			}
			if credential != tt.expectedCredential {
				t.Errorf("credential = %q, want %q", credential, tt.expectedCredential)
			}
		})
	}
}

func TestResolveICEServers(t *testing.T) {
	now := time.Unix(1700000000, 0)

	servers := []types.ICEServer{
		{URLs: []string{"stun:stun.example.com:3478"}},
		{URLs: []string{"turn:static.example.com:3478"}, Username: "neko", Credential: "neko"},
		{URLs: []string{"turn:rest.example.com:3478", "turns:rest.example.com:5349"}, Secret: "north"},
		{URLs: []string{"turn:short.example.com:3478"}, Username: "ignored", Credential: "ignored", Secret: "north", TTL: 3600},
	}
	original := make([]types.ICEServer, len(servers))
	copy(original, servers)

	resolved := resolveICEServers(servers, "alice", now)

	expected := []types.ICEServer{
		{URLs: []string{"stun:stun.example.com:3478"}},
		{URLs: []string{"turn:static.example.com:3478"}, Username: "neko", Credential: "neko"},
		// default TTL of one day
		{URLs: []string{"turn:rest.example.com:3478", "turns:rest.example.com:5349"}, Username: "1700086400:alice", Credential: "SXua5ne/+mDhiHTp0pQJzRO4ESg="},
		// configured TTL replaces the default, generated credentials replace static ones
		{URLs: []string{"turn:short.example.com:3478"}, Username: "1700003600:alice", Credential: "wjwSXO2ch1B6VaLTLMy2Avn5O9o="},
	}
	if !reflect.DeepEqual(resolved, expected) {
		t.Errorf("resolved ICE servers:\n got %+v\nwant %+v", resolved, expected)
	}

	// configuration must stay untouched, so that every connection gets fresh credentials
	if !reflect.DeepEqual(servers, original) {
		t.Errorf("configured ICE servers were modified:\n got %+v\nwant %+v", servers, original)
	}
}

func TestICEServerSecretNotSerialized(t *testing.T) {
	data, err := json.Marshal(types.ICEServer{
		URLs:   []string{"turn:rest.example.com:3478"},
		Secret: "north",
		TTL:    3600,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"north", "secret", "ttl"} {
		if strings.Contains(string(data), s) {
			t.Errorf("serialized ICE server contains %q: %s", s, data)
		}
	}
}

func TestManagerICEServersPerSession(t *testing.T) {
	manager := &WebRTCManagerCtx{
		config: &config.WebRTC{
			ICEServersFrontend: []types.ICEServer{
				{URLs: []string{"turn:rest.example.com:3478"}, Secret: "north"},
			},
			ICEServersBackend: []types.ICEServer{
				{URLs: []string{"stun:stun.example.com:3478"}},
				{URLs: []string{"turn:rest.example.com:3478"}, Secret: "south"},
			},
		},
	}

	// clients get credentials bound to their session
	frontend := manager.ICEServers(&fakeSession{id: "alice"})
	if len(frontend) != 1 || !strings.HasSuffix(frontend[0].Username, ":alice") || frontend[0].Credential == "" || frontend[0].Secret != "" {
		t.Errorf("unexpected frontend ICE servers for alice: %+v", frontend)
	}

	// so do the backend ICE servers of the peer connection
	backend := manager.webrtcConfiguration(&fakeSession{id: "bob"}).ICEServers
	if len(backend) != 2 {
		t.Fatalf("expected 2 backend ICE servers, got %+v", backend)
	}
	if backend[0].Username != "" || backend[0].Credential != false {
		t.Errorf("STUN server must stay without credentials, got %+v", backend[0])
	}
	if !strings.HasSuffix(backend[1].Username, ":bob") || backend[1].Credential == "" || backend[1].CredentialType != webrtc.ICECredentialTypePassword {
		t.Errorf("unexpected backend TURN server for bob: %+v", backend[1])
	}

	// ICE Lite does not use any backend ICE servers
	manager.config.ICELite = true
	if servers := manager.webrtcConfiguration(&fakeSession{id: "bob"}).ICEServers; len(servers) != 0 {
		t.Errorf("expected no backend ICE servers with ICE Lite, got %+v", servers)
	}
}
