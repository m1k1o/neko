package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/api"
	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/member"
	"github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/types"
)

type testWebSocketManager struct{}

func (testWebSocketManager) Start() {}

func (testWebSocketManager) Shutdown() error { return nil }

func (testWebSocketManager) AddHandler(types.WebSocketHandler) {}

func (testWebSocketManager) Upgrade(types.CheckOrigin) types.RouterHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		http.NotFound(w, r)
		return nil
	}
}

func TestOAuthAutoRedirectSkipsAuthenticatedSessions(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(staticDir+"/index.html", []byte("neko client"), 0o644); err != nil {
		t.Fatal(err)
	}

	memberConfig := &config.Member{Provider: "oauth", OAuth: config.OAuth{Enabled: true, AutoRedirect: true}}
	sessionManager := session.New(&config.Session{
		Cookie: config.SessionCookie{Enabled: true, Name: "NEKO_SESSION", Expiration: time.Hour},
	})
	apiManager := api.New(sessionManager, member.New(sessionManager, memberConfig), nil, nil, memberConfig, &config.Server{})
	manager := New(testWebSocketManager{}, apiManager, &config.Server{Static: staticDir}, memberConfig)

	unauthenticated := httptest.NewRecorder()
	manager.router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/", nil))
	if unauthenticated.Code != http.StatusFound {
		t.Fatalf("unauthenticated status = %d", unauthenticated.Code)
	}
	if location := unauthenticated.Header().Get("Location"); location != "/api/oauth/login" {
		t.Fatalf("unauthenticated location = %q", location)
	}

	_, token, err := sessionManager.Create("oauth:user-123", types.MemberProfile{CanLogin: true})
	if err != nil {
		t.Fatal(err)
	}
	authenticatedRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	authenticatedRequest.AddCookie(&http.Cookie{Name: "NEKO_SESSION", Value: token})
	authenticated := httptest.NewRecorder()
	manager.router.ServeHTTP(authenticated, authenticatedRequest)
	if authenticated.Code != http.StatusOK {
		t.Fatalf("authenticated status = %d", authenticated.Code)
	}
	if body := authenticated.Body.String(); body != "neko client" {
		t.Fatalf("authenticated body = %q", body)
	}
}
