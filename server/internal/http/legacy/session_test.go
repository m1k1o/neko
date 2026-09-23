package legacy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestCreateUsesExistingSessionCookie(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/whoami":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if got := r.Header.Get("Cookie"); got != "NEKO_SESSION=session-token" {
				t.Fatalf("cookie = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "oauth:user-123",
				"profile": types.MemberProfile{
					Name:    "Ada Lovelace",
					IsAdmin: true,
				},
			})
		case "/api/protected":
			if got := r.Header.Get("Cookie"); got != "NEKO_SESSION=session-token" {
				t.Fatalf("authenticated request cookie = %q", got)
			}
			_ = json.NewEncoder(w).Encode(true)
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	request := httptest.NewRequest(http.MethodGet, "http://neko.test/ws", nil)
	request.Header.Set("Cookie", "NEKO_SESSION=session-token")
	handler := &LegacyHandler{sessionIPs: make(map[string]string)}
	session := &session{
		r:          request,
		h:          handler,
		serverAddr: strings.TrimPrefix(backend.URL, "http://"),
		client:     backend.Client(),
	}

	if err := session.create("", ""); err != nil {
		t.Fatal(err)
	}
	if session.id != "oauth:user-123" || session.name != "Ada Lovelace" || !session.isAdmin {
		t.Fatalf("session = %#v", session)
	}
	if session.cookie != "NEKO_SESSION=session-token" || session.token != "" {
		t.Fatalf("authentication state = cookie %q token %q", session.cookie, session.token)
	}
	if err := session.apiReq(http.MethodGet, "/api/protected", nil, nil); err != nil {
		t.Fatal(err)
	}
}
