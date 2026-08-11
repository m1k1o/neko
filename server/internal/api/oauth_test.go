package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/member"
	"github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/types"
)

func TestOAuthLoginSynchronizesProfile(t *testing.T) {
	var tokenRequest url.Values
	idToken := testIDToken(t, map[string]string{
		"sub":         "user-123",
		"displayName": "Ada Lovelace",
		"avatar":      "https://example.test/ada.png",
		"isAdmin":     "true",
	})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tokenRequest = r.Form
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "provider-token", "id_token": idToken})
		case "/userinfo":
			if got := r.Header.Get("Authorization"); got != "Bearer provider-token" {
				http.Error(w, "missing provider authorization", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{
				"sub":   "user-123",
				"email": "user@example.test",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()

	memberConfig := &config.Member{
		Provider: "oauth",
		OAuth: config.OAuth{
			Enabled:          true,
			ClientID:         "client-id",
			ClientSecret:     "client-secret",
			AuthorizationURL: provider.URL + "/authorize",
			TokenURL:         provider.URL + "/token",
			UserInfoURL:      provider.URL + "/userinfo",
			RedirectURL:      "https://neko.example.test/api/oauth/callback",
			Scopes:           []string{"openid", "profile"},
			SubjectField:     "sub",
			UsernameField:    "displayName",
			AvatarField:      "avatar",
			AdminEmails:      []string{"admin@example.test"},
			SuccessRedirect:  "/room",
			UserProfile: types.MemberProfile{
				CanLogin: true,
			},
		},
	}
	sessionManager := session.New(&config.Session{
		Cookie: config.SessionCookie{Enabled: true, Name: "NEKO_SESSION", Expiration: time.Hour},
	})
	members := member.New(sessionManager, memberConfig)
	api := New(sessionManager, members, nil, nil, memberConfig, &config.Server{PathPrefix: "/"})

	loginRecorder := httptest.NewRecorder()
	api.OAuthLogin(loginRecorder, httptest.NewRequest(http.MethodGet, "/api/oauth/login", nil))
	if loginRecorder.Code != http.StatusFound {
		t.Fatalf("login status = %d", loginRecorder.Code)
	}
	authorizeURL, err := url.Parse(loginRecorder.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	state := authorizeURL.Query().Get("state")
	if state == "" || authorizeURL.Query().Get("code_challenge") == "" {
		t.Fatalf("OAuth redirect is missing state or PKCE challenge: %s", authorizeURL.String())
	}

	callbackRecorder := httptest.NewRecorder()
	callbackURL := "/api/oauth/callback?state=" + url.QueryEscape(state) + "&code=authorization-code"
	api.OAuthCallback(callbackRecorder, httptest.NewRequest(http.MethodGet, callbackURL, nil))
	if callbackRecorder.Code != http.StatusSeeOther {
		t.Fatalf("callback status = %d", callbackRecorder.Code)
	}
	if got := callbackRecorder.Header().Get("Location"); got != "/room" {
		t.Fatalf("redirect = %q", got)
	}
	if tokenRequest.Get("code_verifier") == "" {
		t.Fatal("token request is missing PKCE verifier")
	}
	cookies := callbackRecorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("OAuth callback did not set a session cookie")
	}
	if cookies[0].Path != "/" {
		t.Fatalf("OAuth session cookie path = %q, want root path", cookies[0].Path)
	}

	userSession, ok := sessionManager.Get("oauth:user-123")
	if !ok {
		t.Fatal("OAuth session was not created")
	}
	profile := userSession.Profile()
	if profile.Name != "Ada Lovelace" || profile.Avatar != "https://example.test/ada.png" || !profile.IsAdmin {
		t.Fatalf("profile = %#v", profile)
	}
	if extraData := members.OAuthExtraData(userSession.ID()); extraData["displayName"] != "Ada Lovelace" || extraData["isAdmin"] != "true" {
		t.Fatalf("extra data = %#v", extraData)
	}
}

func testIDToken(t *testing.T, claims map[string]string) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return "eyJhbGciOiJub25lIn0." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func TestOAuthIssuerDiscoveryTakesPrecedence(t *testing.T) {
	var tokenRequest url.Values
	var issuerURL string
	issuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 issuerURL,
				"authorization_endpoint": issuerURL + "/authorize",
				"token_endpoint":         issuerURL + "/token",
				"userinfo_endpoint":      issuerURL + "/userinfo",
			})
		case "/token":
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tokenRequest = r.Form
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "provider-token"})
		case "/userinfo":
			_ = json.NewEncoder(w).Encode(map[string]string{"sub": "user-456", "name": "Grace Hopper"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer issuer.Close()

	issuerURL = issuer.URL
	memberConfig := &config.Member{
		Provider: "oauth",
		OAuth: config.OAuth{
			Enabled:          true,
			ClientID:         "client-id",
			ClientSecret:     "client-secret",
			IssuerURL:        issuerURL,
			AuthorizationURL: "https://ignored.example.test/authorize",
			TokenURL:         "https://ignored.example.test/token",
			UserInfoURL:      "https://ignored.example.test/userinfo",
			Scopes:           []string{"openid", "profile"},
			SubjectField:     "sub",
			UsernameField:    "name",
			AvatarField:      "picture",
			UserProfile:      types.MemberProfile{CanLogin: true},
		},
	}

	sessionManager := session.New(&config.Session{
		Cookie: config.SessionCookie{Enabled: true, Name: "NEKO_SESSION", Expiration: time.Hour},
	})
	members := member.New(sessionManager, memberConfig)
	api := New(sessionManager, members, nil, nil, memberConfig, &config.Server{PathPrefix: "/"})

	loginRequest := httptest.NewRequest(http.MethodGet, "https://neko.example.test/api/oauth/login", nil)
	loginRecorder := httptest.NewRecorder()
	api.OAuthLogin(loginRecorder, loginRequest)
	if loginRecorder.Code != http.StatusFound {
		t.Fatalf("login status = %d", loginRecorder.Code)
	}
	authorizationURL, err := url.Parse(loginRecorder.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	issuerEndpoint, err := url.Parse(issuerURL)
	if err != nil {
		t.Fatal(err)
	}
	if authorizationURL.Host != issuerEndpoint.Host {
		t.Fatalf("authorization endpoint = %q", authorizationURL.String())
	}
	redirectURL := authorizationURL.Query().Get("redirect_uri")
	if redirectURL != "https://neko.example.test/api/oauth/callback" {
		t.Fatalf("derived redirect URL = %q", redirectURL)
	}

	callbackRecorder := httptest.NewRecorder()
	callbackURL := "/api/oauth/callback?state=" + url.QueryEscape(authorizationURL.Query().Get("state")) + "&code=authorization-code"
	api.OAuthCallback(callbackRecorder, httptest.NewRequest(http.MethodGet, callbackURL, nil))
	if callbackRecorder.Code != http.StatusSeeOther {
		t.Fatalf("callback status = %d", callbackRecorder.Code)
	}
	if tokenRequest.Get("redirect_uri") != redirectURL {
		t.Fatalf("token redirect URL = %q, want %q", tokenRequest.Get("redirect_uri"), redirectURL)
	}
}

func TestOAuthConfigUsesProviderAndName(t *testing.T) {
	memberConfig := &config.Member{
		Provider: "oauth",
		OAuth: config.OAuth{
			Enabled: true,
			Name:    "Team SSO",
		},
	}
	sessionManager := session.New(&config.Session{})
	api := New(sessionManager, member.New(sessionManager, memberConfig), nil, nil, memberConfig, &config.Server{})

	recorder := httptest.NewRecorder()
	if err := api.OAuthConfig(recorder, httptest.NewRequest(http.MethodGet, "/api/oauth/config", nil)); err != nil {
		t.Fatal(err)
	}

	var payload OAuthUIConfig
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Enabled || payload.PasswordLoginEnabled || payload.Name != "Team SSO" || payload.LoginURL != "/api/oauth/login" {
		t.Fatalf("OAuth UI config = %#v", payload)
	}
}
