package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/member"
	"github.com/m1k1o/neko/server/internal/member/oauth"
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
		OAuth: oauth.Config{
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
				CanLogin:           true,
				CanAccessClipboard: true,
			},
			AdminProfile: types.MemberProfile{
				IsAdmin:               true,
				CanLogin:              true,
				CanSeeInactiveCursors: true,
			},
		},
	}
	sessionManager := session.New(&config.Session{
		Cookie: config.SessionCookie{Enabled: true, Name: "NEKO_SESSION", Expiration: time.Hour},
	})
	members := member.New(sessionManager, memberConfig)
	api := New(sessionManager, members, nil, nil, memberConfig, &config.Server{PathPrefix: "/"})

	loginRecorder := httptest.NewRecorder()
	api.OAuthLogin(loginRecorder, httptest.NewRequest(http.MethodGet, "https://neko.example.test/api/oauth/login", nil))
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
	stateCookie := requireOAuthStateCookie(t, loginRecorder, state)
	if stateCookie.Value != state {
		t.Fatal("OAuth state cookie does not match authorization state")
	}

	callbackRecorder := httptest.NewRecorder()
	callbackURL := "/api/oauth/callback?state=" + url.QueryEscape(state) + "&code=authorization-code"
	callbackRequest := httptest.NewRequest(http.MethodGet, callbackURL, nil)
	callbackRequest.AddCookie(stateCookie)
	if err := api.OAuthCallback(callbackRecorder, callbackRequest); err != nil {
		t.Fatal(err)
	}
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
	sessionCookie := findCookie(cookies, "NEKO_SESSION")
	if sessionCookie == nil {
		t.Fatal("OAuth callback did not set a session cookie")
	}
	if sessionCookie.Path != "/" {
		t.Fatalf("OAuth session cookie path = %q, want root path", sessionCookie.Path)
	}
	clearedStateCookie := findCookie(cookies, oauthStateCookieName(state))
	if clearedStateCookie == nil || clearedStateCookie.MaxAge >= 0 {
		t.Fatal("OAuth callback did not clear the state cookie")
	}

	userSession, ok := sessionManager.Get("oauth:user-123")
	if !ok {
		t.Fatal("OAuth session was not created")
	}
	profile := userSession.Profile()
	if profile.Name != "Ada Lovelace" || profile.Avatar != "https://example.test/ada.png" || !profile.IsAdmin || profile.CanAccessClipboard {
		t.Fatalf("profile = %#v", profile)
	}
	if extraData := api.oauth.service.ExtraData(userSession.ID()); extraData["displayName"] != "Ada Lovelace" || extraData["isAdmin"] != "true" {
		t.Fatalf("extra data = %#v", extraData)
	}
}

func requireOAuthStateCookie(t *testing.T, recorder *httptest.ResponseRecorder, stateToken string) *http.Cookie {
	t.Helper()

	cookie := findCookie(recorder.Result().Cookies(), oauthStateCookieName(stateToken))
	if cookie == nil {
		t.Fatal("OAuth login did not set a state cookie")
	}
	if cookie.Path != "/" {
		t.Fatalf("OAuth state cookie path = %q", cookie.Path)
	}
	if cookie.MaxAge != int(oauth.StateLifetime/time.Second) {
		t.Fatalf("OAuth state cookie max age = %d", cookie.MaxAge)
	}
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("OAuth state cookie flags = Secure:%t HttpOnly:%t SameSite:%d", cookie.Secure, cookie.HttpOnly, cookie.SameSite)
	}
	return cookie
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func TestOAuthCallbackRequiresStateCookie(t *testing.T) {
	tokenRequests := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			tokenRequests++
		}
		http.Error(w, "invalid request", http.StatusBadRequest)
	}))
	defer provider.Close()

	memberConfig := &config.Member{
		Provider: "oauth",
		OAuth: oauth.Config{
			Enabled:          true,
			ClientID:         "client-id",
			ClientSecret:     "client-secret",
			AuthorizationURL: provider.URL + "/authorize",
			TokenURL:         provider.URL + "/token",
			UserInfoURL:      provider.URL + "/userinfo",
			RedirectURL:      "https://neko.example.test/api/oauth/callback",
			Scopes:           []string{"openid"},
			SubjectField:     "sub",
			UsernameField:    "name",
			UserProfile:      types.MemberProfile{CanLogin: true},
		},
	}
	sessionManager := session.New(&config.Session{
		Cookie: config.SessionCookie{Enabled: true, Name: "NEKO_SESSION", Expiration: time.Hour},
	})
	members := member.New(sessionManager, memberConfig)
	api := New(sessionManager, members, nil, nil, memberConfig, &config.Server{PathPrefix: "/"})

	loginRecorder := httptest.NewRecorder()
	if err := api.OAuthLogin(loginRecorder, httptest.NewRequest(http.MethodGet, "https://neko.example.test/api/oauth/login", nil)); err != nil {
		t.Fatal(err)
	}
	authorizationURL, err := url.Parse(loginRecorder.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	state := authorizationURL.Query().Get("state")
	if state == "" {
		t.Fatal("OAuth redirect is missing state")
	}

	callbackURL := "/api/oauth/callback?state=" + url.QueryEscape(state) + "&code=authorization-code"
	callbackErr := api.OAuthCallback(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, callbackURL, nil))
	if callbackErr == nil {
		t.Fatal("callback without the initiating browser cookie succeeded")
	}
	if tokenRequests != 0 {
		t.Fatalf("callback without the initiating browser cookie reached token endpoint %d times", tokenRequests)
	}

	wrongCookieRequest := httptest.NewRequest(http.MethodGet, callbackURL, nil)
	wrongCookieRequest.AddCookie(&http.Cookie{Name: oauthStateCookieName(state), Value: "wrong-state"})
	if err := api.OAuthCallback(httptest.NewRecorder(), wrongCookieRequest); err == nil {
		t.Fatal("callback with a mismatched state cookie succeeded")
	}
	if tokenRequests != 0 {
		t.Fatalf("callback with a mismatched state cookie reached token endpoint %d times", tokenRequests)
	}

	firstCookie := requireOAuthStateCookie(t, loginRecorder, state)
	secondLoginRecorder := httptest.NewRecorder()
	if err := api.OAuthLogin(secondLoginRecorder, httptest.NewRequest(http.MethodGet, "https://neko.example.test/api/oauth/login", nil)); err != nil {
		t.Fatal(err)
	}
	secondAuthorizationURL, err := url.Parse(secondLoginRecorder.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	secondState := secondAuthorizationURL.Query().Get("state")
	secondCookie := requireOAuthStateCookie(t, secondLoginRecorder, secondState)
	if firstCookie.Name == secondCookie.Name {
		t.Fatal("concurrent OAuth logins use the same state cookie name")
	}

	concurrentRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	concurrentRequest.AddCookie(firstCookie)
	concurrentRequest.AddCookie(secondCookie)
	if !api.oauth.stateCookieMatches(concurrentRequest, state) || !api.oauth.stateCookieMatches(concurrentRequest, secondState) {
		t.Fatal("concurrent OAuth state cookies are not independently valid")
	}

	declinedRecorder := httptest.NewRecorder()
	declinedURL := "/api/oauth/callback?state=" + url.QueryEscape(secondState) + "&error=access_denied"
	declinedRequest := httptest.NewRequest(http.MethodGet, declinedURL, nil)
	declinedRequest.AddCookie(secondCookie)
	if err := api.OAuthCallback(declinedRecorder, declinedRequest); err == nil {
		t.Fatal("declined OAuth callback succeeded")
	}
	clearedCookie := findCookie(declinedRecorder.Result().Cookies(), secondCookie.Name)
	if clearedCookie == nil || clearedCookie.MaxAge >= 0 {
		t.Fatal("declined OAuth callback did not clear its state cookie")
	}
}

func TestOAuthLoginUsesConfiguredRedirectBehindProxy(t *testing.T) {
	memberConfig := &config.Member{
		Provider: "oauth",
		OAuth: oauth.Config{
			Enabled:          true,
			ClientID:         "client-id",
			ClientSecret:     "client-secret",
			AuthorizationURL: "https://auth.example.test/authorize",
			TokenURL:         "https://auth.example.test/token",
			UserInfoURL:      "https://auth.example.test/userinfo",
			RedirectURL:      "https://neko.example.test/api/oauth/callback",
			Scopes:           []string{"openid"},
			SubjectField:     "sub",
			UsernameField:    "name",
			UserProfile:      types.MemberProfile{CanLogin: true},
		},
	}
	sessionManager := session.New(&config.Session{
		Cookie: config.SessionCookie{Enabled: true, Name: "NEKO_SESSION", Expiration: time.Hour},
	})
	members := member.New(sessionManager, memberConfig)
	api := New(sessionManager, members, nil, nil, memberConfig, &config.Server{PathPrefix: "/"})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://neko.example.test/api/oauth/login", nil)
	if err := api.OAuthLogin(recorder, request); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusFound {
		t.Fatalf("login status = %d", recorder.Code)
	}

	authorizationURL, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if got := authorizationURL.Query().Get("redirect_uri"); got != memberConfig.OAuth.RedirectURL {
		t.Fatalf("OAuth redirect_uri = %q", got)
	}
	state := authorizationURL.Query().Get("state")
	cookie := requireOAuthStateCookie(t, recorder, state)
	if !cookie.Secure {
		t.Fatal("HTTPS OAuth callback requires a Secure state cookie")
	}
}

func TestOAuthStateCookieValidatesRedirectHost(t *testing.T) {
	handler := newOAuthHandler(nil, "/", false, true)
	location := func(redirectURL string) string {
		values := url.Values{"state": {"state-token"}, "redirect_uri": {redirectURL}}
		return "https://auth.example.test/authorize?" + values.Encode()
	}

	mismatchRecorder := httptest.NewRecorder()
	if err := handler.setStateCookie(
		mismatchRecorder,
		location("https://other.example.test/api/oauth/callback"),
		"https://neko.example.test/api/oauth/callback",
	); err == nil {
		t.Fatal("OAuth redirect host mismatch was accepted")
	}
	if len(mismatchRecorder.Result().Cookies()) != 0 {
		t.Fatal("OAuth redirect host mismatch set a state cookie")
	}

	publicPathRecorder := httptest.NewRecorder()
	publicCallback := "https://neko.example.test/public/oauth/callback"
	if err := handler.setStateCookie(
		publicPathRecorder,
		location(publicCallback),
		"https://neko.example.test:443/api/oauth/callback",
	); err != nil {
		t.Fatal(err)
	}
	cookie := requireOAuthStateCookie(t, publicPathRecorder, "state-token")
	if cookie.Path != "/" {
		t.Fatalf("OAuth state cookie path = %q", cookie.Path)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	loginURL, _ := url.Parse("https://neko.example.test/api/oauth/login")
	callbackURL, _ := url.Parse(publicCallback)
	jar.SetCookies(loginURL, publicPathRecorder.Result().Cookies())
	if findCookie(jar.Cookies(callbackURL), cookie.Name) == nil {
		t.Fatal("browser cookie jar did not send the state cookie to the configured callback path")
	}

	httpRecorder := httptest.NewRecorder()
	if err := handler.setStateCookie(
		httpRecorder,
		location("http://neko.example.test/api/oauth/callback"),
		"http://neko.example.test/api/oauth/callback",
	); err != nil {
		t.Fatal(err)
	}
	if cookie := findCookie(httpRecorder.Result().Cookies(), oauthStateCookieName("state-token")); cookie == nil || cookie.Secure {
		t.Fatal("HTTP OAuth state cookie must not be Secure")
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
		OAuth: oauth.Config{
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
	state := authorizationURL.Query().Get("state")
	callbackURL := "/api/oauth/callback?state=" + url.QueryEscape(state) + "&code=authorization-code"
	callbackRequest := httptest.NewRequest(http.MethodGet, callbackURL, nil)
	callbackRequest.AddCookie(requireOAuthStateCookie(t, loginRecorder, state))
	if err := api.OAuthCallback(callbackRecorder, callbackRequest); err != nil {
		t.Fatal(err)
	}
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
		OAuth: oauth.Config{
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
