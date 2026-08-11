package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

const oauthStateLifetime = 10 * time.Minute

type oauthState struct {
	verifier    string
	redirectURL string
	expires     time.Time
}

type oauthToken struct {
	accessToken   string
	idTokenClaims map[string]any
}

type OAuthUIConfig struct {
	Enabled              bool   `json:"enabled"`
	Name                 string `json:"name"`
	LoginURL             string `json:"login_url"`
	PasswordLoginEnabled bool   `json:"password_login_enabled"`
}

type oauthHandler struct {
	config         config.OAuth
	client         *http.Client
	pathPrefix     string
	trustProxy     bool
	providerOAuth  bool
	configMu       sync.Mutex
	issuerResolved bool
	statesMu       sync.Mutex
	states         map[string]oauthState
}

func newOAuthHandler(config config.OAuth, pathPrefix string, trustProxy bool, providerOAuth bool) *oauthHandler {
	if pathPrefix == "" {
		pathPrefix = "/"
	}
	return &oauthHandler{
		config:        config,
		client:        &http.Client{Timeout: 15 * time.Second},
		pathPrefix:    pathPrefix,
		trustProxy:    trustProxy,
		providerOAuth: providerOAuth,
		states:        make(map[string]oauthState),
	}
}

func (handler *oauthHandler) configured(config config.OAuth) bool {
	return handler.providerOAuth && config.Enabled && config.ClientID != "" && config.ClientSecret != "" &&
		config.AuthorizationURL != "" && config.TokenURL != "" && config.UserInfoURL != ""
}

func (api *ApiManagerCtx) OAuthConfig(w http.ResponseWriter, r *http.Request) error {
	config := api.oauth.config
	name := strings.TrimSpace(config.Name)
	if name == "" {
		name = "OAuth"
	}

	return utils.HttpSuccess(w, OAuthUIConfig{
		Enabled:              api.oauth.providerOAuth && config.Enabled,
		Name:                 name,
		LoginURL:             path.Join(api.oauth.pathPrefix, "/api/oauth/login"),
		PasswordLoginEnabled: !api.oauth.providerOAuth,
	})
}

type openIDConfiguration struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

// resolveConfig uses OpenID Connect discovery when an issuer is configured.
// Issuer metadata deliberately takes precedence over explicitly configured endpoints.
func (handler *oauthHandler) resolveConfig(ctx context.Context) (config.OAuth, error) {
	handler.configMu.Lock()
	defer handler.configMu.Unlock()

	if handler.config.IssuerURL == "" || handler.issuerResolved {
		return handler.config, nil
	}

	discoveryURL, err := openIDConfigurationURL(handler.config.IssuerURL)
	if err != nil {
		return config.OAuth{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return config.OAuth{}, err
	}
	req.Header.Set("Accept", "application/json")
	response, err := handler.client.Do(req)
	if err != nil {
		return config.OAuth{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return config.OAuth{}, fmt.Errorf("issuer discovery endpoint returned %s", response.Status)
	}

	var metadata openIDConfiguration
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&metadata); err != nil {
		return config.OAuth{}, err
	}
	if metadata.Issuer != handler.config.IssuerURL {
		return config.OAuth{}, fmt.Errorf("issuer discovery response issuer does not match configured issuer")
	}
	for field, endpoint := range map[string]string{
		"authorization_endpoint": metadata.AuthorizationEndpoint,
		"token_endpoint":         metadata.TokenEndpoint,
		"userinfo_endpoint":      metadata.UserInfoEndpoint,
	} {
		if !isAbsoluteHTTPURL(endpoint) {
			return config.OAuth{}, fmt.Errorf("issuer discovery response has invalid %s", field)
		}
	}

	handler.config.AuthorizationURL = metadata.AuthorizationEndpoint
	handler.config.TokenURL = metadata.TokenEndpoint
	handler.config.UserInfoURL = metadata.UserInfoEndpoint
	handler.issuerResolved = true
	return handler.config, nil
}

func openIDConfigurationURL(issuerURL string) (string, error) {
	issuer, err := url.Parse(issuerURL)
	if err != nil || !isAbsoluteHTTPURL(issuerURL) {
		return "", errors.New("OAuth issuer URL must be an absolute HTTP(S) URL")
	}
	issuer.Path = strings.TrimSuffix(issuer.Path, "/") + "/.well-known/openid-configuration"
	issuer.RawQuery = ""
	issuer.Fragment = ""
	return issuer.String(), nil
}

func isAbsoluteHTTPURL(value string) bool {
	endpoint, err := url.Parse(value)
	return err == nil && endpoint.Host != "" && (endpoint.Scheme == "http" || endpoint.Scheme == "https")
}

func (handler *oauthHandler) callbackURL(r *http.Request) (string, error) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if handler.trustProxy {
		if forwardedProto := forwardedHeaderValue(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
			scheme = forwardedProto
		}
		if forwardedHost := forwardedHeaderValue(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
			host = forwardedHost
		}
	}
	if host == "" || (scheme != "http" && scheme != "https") {
		return "", errors.New("request has no valid public URL")
	}

	callback := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path.Join(handler.pathPrefix, "/api/oauth/callback"),
	}
	if !isAbsoluteHTTPURL(callback.String()) {
		return "", errors.New("request has no valid public URL")
	}
	return callback.String(), nil
}

func forwardedHeaderValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(value, ",")[0])
}

func (api *ApiManagerCtx) OAuthLogin(w http.ResponseWriter, r *http.Request) error {
	if !api.oauth.providerOAuth || !api.oauth.config.Enabled {
		return utils.HttpNotFound()
	}
	config, err := api.oauth.resolveConfig(r.Context())
	if err != nil {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth issuer discovery failed").WithInternalErr(err)
	}
	if !api.oauth.configured(config) {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth is not fully configured")
	}
	if !api.sessions.CookieEnabled() {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth requires session cookies to be enabled")
	}

	state, err := newOAuthToken(32)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}
	verifier, err := newOAuthToken(64)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	redirectURL := config.RedirectURL
	if redirectURL == "" {
		redirectURL, err = api.oauth.callbackURL(r)
		if err != nil {
			return utils.HttpInternalServerError("unable to determine OAuth callback URL").WithInternalErr(err)
		}
	}

	api.oauth.storeState(state, verifier, redirectURL)

	authorizationURL, err := url.Parse(config.AuthorizationURL)
	if err != nil {
		return utils.HttpInternalServerError("invalid OAuth authorization URL").WithInternalErr(err)
	}
	query := authorizationURL.Query()
	query.Set("response_type", "code")
	query.Set("client_id", config.ClientID)
	query.Set("redirect_uri", redirectURL)
	query.Set("scope", strings.Join(config.Scopes, " "))
	query.Set("state", state)
	query.Set("code_challenge", oauthChallenge(verifier))
	query.Set("code_challenge_method", "S256")
	authorizationURL.RawQuery = query.Encode()

	http.Redirect(w, r, authorizationURL.String(), http.StatusFound)
	return nil
}

func (api *ApiManagerCtx) OAuthCallback(w http.ResponseWriter, r *http.Request) error {
	if !api.oauth.providerOAuth || !api.oauth.config.Enabled {
		return utils.HttpNotFound()
	}
	config, err := api.oauth.resolveConfig(r.Context())
	if err != nil {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth issuer discovery failed").WithInternalErr(err)
	}
	if !api.oauth.configured(config) {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth is not fully configured")
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		return utils.HttpUnauthorized("OAuth authorization was declined").WithInternalMsg(providerError)
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		return utils.HttpBadRequest("OAuth callback is missing state or code")
	}
	oauthState, ok := api.oauth.popState(state)
	if !ok {
		return utils.HttpBadRequest("OAuth state is invalid or expired")
	}

	oauthToken, err := api.oauth.exchangeCode(r, config, code, oauthState.verifier, oauthState.redirectURL)
	if err != nil {
		return utils.HttpUnauthorized("OAuth token exchange failed").WithInternalErr(err)
	}
	claims, err := api.oauth.userInfo(r, config, oauthToken.accessToken)
	if err != nil {
		return utils.HttpUnauthorized("OAuth user-info request failed").WithInternalErr(err)
	}

	subject := oauthClaim(claims, config.SubjectField)
	if subject == "" {
		return utils.HttpUnauthorized("OAuth user-info response is missing the subject field")
	}
	if idTokenSubject := oauthClaim(oauthToken.idTokenClaims, config.SubjectField); idTokenSubject != "" && idTokenSubject != subject {
		return utils.HttpUnauthorized("OAuth ID token subject does not match user-info response")
	}
	name := oauthClaim(claims, config.UsernameField)
	if name == "" {
		name = oauthClaim(oauthToken.idTokenClaims, config.UsernameField)
	}
	if name == "" {
		name = subject
	}
	avatar := oauthClaim(claims, config.AvatarField)
	if avatar == "" {
		avatar = oauthClaim(oauthToken.idTokenClaims, config.AvatarField)
	}

	email := oauthClaim(claims, "email")
	if email == "" {
		email = oauthClaim(oauthToken.idTokenClaims, "email")
	}
	isAdmin := oauthBoolClaim(claims, "isAdmin") || oauthBoolClaim(oauthToken.idTokenClaims, "isAdmin")
	_, token, err := api.members.LoginOAuth(subject, name, avatar, email, isAdmin, mergeOAuthClaims(claims, oauthToken.idTokenClaims))
	if err != nil {
		if errors.Is(err, types.ErrSessionAlreadyConnected) {
			return utils.HttpUnprocessableEntity("session already connected")
		}
		if errors.Is(err, types.ErrSessionLoginsLocked) {
			return utils.HttpForbidden("logins are locked").WithInternalErr(err)
		}
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	api.sessions.CookieSetToken(w, token)
	http.Redirect(w, r, api.oauth.successRedirect(), http.StatusSeeOther)
	return nil
}

func (handler *oauthHandler) storeState(state, verifier, redirectURL string) {
	handler.statesMu.Lock()
	defer handler.statesMu.Unlock()

	now := time.Now()
	for key, value := range handler.states {
		if now.After(value.expires) {
			delete(handler.states, key)
		}
	}
	handler.states[state] = oauthState{verifier: verifier, redirectURL: redirectURL, expires: now.Add(oauthStateLifetime)}
}

func (handler *oauthHandler) popState(state string) (oauthState, bool) {
	handler.statesMu.Lock()
	defer handler.statesMu.Unlock()

	value, ok := handler.states[state]
	delete(handler.states, state)
	return value, ok && time.Now().Before(value.expires)
}

func (handler *oauthHandler) exchangeCode(r *http.Request, config config.OAuth, code, verifier, redirectURL string) (oauthToken, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURL},
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"code_verifier": {verifier},
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, config.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return oauthToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	response, err := handler.client.Do(req)
	if err != nil {
		return oauthToken{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return oauthToken{}, fmt.Errorf("token endpoint returned %s", response.Status)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return oauthToken{}, err
	}
	if payload.AccessToken == "" {
		return oauthToken{}, errors.New("token response has no access_token")
	}

	return oauthToken{
		accessToken:   payload.AccessToken,
		idTokenClaims: idTokenClaims(payload.IDToken),
	}, nil
}

// idTokenClaims is used only as a profile fallback when userinfo omits optional
// OIDC claims. The userinfo response remains the source of the session subject.
func idTokenClaims(idToken string) map[string]any {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}
	return claims
}

func (handler *oauthHandler) userInfo(r *http.Request, config config.OAuth, accessToken string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := handler.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("user-info endpoint returned %s", response.Status)
	}

	claims := map[string]any{}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func (handler *oauthHandler) successRedirect() string {
	redirect := handler.config.SuccessRedirect
	if strings.HasPrefix(redirect, "/") && !strings.HasPrefix(redirect, "//") {
		return path.Join(handler.pathPrefix, redirect)
	}
	return handler.pathPrefix
}

func newOAuthToken(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func oauthChallenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func oauthClaim(claims map[string]any, field string) string {
	value, ok := claims[field]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}

func oauthBoolClaim(claims map[string]any, field string) bool {
	value, ok := claims[field]
	if !ok {
		return false
	}
	switch value := value.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	case float64:
		return value != 0
	default:
		return false
	}
}

func mergeOAuthClaims(userInfo, idToken map[string]any) map[string]any {
	claims := make(map[string]any, len(userInfo)+len(idToken))
	for key, value := range idToken {
		claims[key] = value
	}
	for key, value := range userInfo {
		claims[key] = value
	}
	return claims
}
