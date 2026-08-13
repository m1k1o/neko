package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

const stateLifetime = 10 * time.Minute

type state struct {
	verifier, redirectURL string
	expires               time.Time
}

// Service owns the OAuth authorization-code flow and creates sessions for the
// OAuth member provider. HTTP packages only translate requests and responses.
type Service struct {
	config         Config
	sessions       types.SessionManager
	client         *http.Client
	mu             sync.Mutex
	issuerResolved bool
	states         map[string]state
	extra          map[string]map[string]any
}

func NewService(config Config, sessions types.SessionManager) *Service {
	service := &Service{
		config:   config,
		sessions: sessions,
		client:   &http.Client{Timeout: 15 * time.Second},
		states:   make(map[string]state),
		extra:    make(map[string]map[string]any),
	}

	sessions.OnDeleted(func(session types.Session) {
		service.deleteExtraData(session.ID())
	})

	return service
}

func (service *Service) Enabled() bool { return service.config.Enabled }

func (service *Service) Name() string {
	if name := strings.TrimSpace(service.config.Name); name != "" {
		return name
	}
	return "OAuth"
}

func (service *Service) SuccessRedirect() string { return service.config.SuccessRedirect }

func (service *Service) Start(ctx context.Context, callbackURL string) (string, error) {
	config, err := service.resolveConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("issuer discovery: %w", err)
	}
	if !config.IsConfigured() {
		return "", errors.New("OAuth is not fully configured")
	}

	if config.RedirectURL != "" {
		callbackURL = config.RedirectURL
	}
	if callbackURL == "" {
		return "", errors.New("OAuth callback URL is not configured")
	}

	stateToken, err := token(32)
	if err != nil {
		return "", err
	}
	verifier, err := token(64)
	if err != nil {
		return "", err
	}

	service.storeState(stateToken, verifier, callbackURL)

	authorizationURL, err := url.Parse(config.AuthorizationURL)
	if err != nil {
		return "", err
	}

	query := authorizationURL.Query()
	query.Set("response_type", "code")
	query.Set("client_id", config.ClientID)
	query.Set("redirect_uri", callbackURL)
	query.Set("scope", strings.Join(config.Scopes, " "))
	query.Set("state", stateToken)
	query.Set("code_challenge", challenge(verifier))
	query.Set("code_challenge_method", "S256")
	authorizationURL.RawQuery = query.Encode()

	return authorizationURL.String(), nil
}

func (service *Service) Complete(ctx context.Context, stateToken, code string) (types.Session, string, error) {
	if !service.config.Enabled {
		return nil, "", types.ErrMemberInvalidPassword
	}

	if stateToken == "" || code == "" {
		return nil, "", errors.New("OAuth callback is missing state or code")
	}

	stored, ok := service.popState(stateToken)
	if !ok {
		return nil, "", errors.New("OAuth state is invalid or expired")
	}

	config, err := service.resolveConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("issuer discovery: %w", err)
	}
	if !config.IsConfigured() {
		return nil, "", errors.New("OAuth is not fully configured")
	}

	accessToken, idClaims, err := service.exchangeCode(ctx, config, code, stored.verifier, stored.redirectURL)
	if err != nil {
		return nil, "", fmt.Errorf("token exchange: %w", err)
	}

	claims, err := service.userInfo(ctx, config, accessToken)
	if err != nil {
		return nil, "", fmt.Errorf("user-info: %w", err)
	}

	subject := claim(claims, config.SubjectField)
	if subject == "" {
		return nil, "", errors.New("OAuth user-info response is missing the subject field")
	}

	if idSubject := claim(idClaims, config.SubjectField); idSubject != "" && idSubject != subject {
		return nil, "", errors.New("OAuth ID token subject does not match user-info response")
	}

	name := claim(claims, config.UsernameField)
	if name == "" {
		name = claim(idClaims, config.UsernameField)
	}
	if name == "" {
		name = subject
	}

	avatar := claim(claims, config.AvatarField)
	if avatar == "" {
		avatar = claim(idClaims, config.AvatarField)
	}

	email := claim(claims, "email")
	if email == "" {
		email = claim(idClaims, "email")
	}

	isAdmin := boolClaim(claims, "isAdmin") || boolClaim(idClaims, "isAdmin")

	var profile types.MemberProfile
	if administrator(email, isAdmin, config.AdminEmails) {
		profile = config.AdminProfile
	} else {
		profile = config.UserProfile
	}

	profile.Name = name
	profile.Avatar = avatar

	if !profile.IsAdmin && service.sessions.Settings().LockedLogins {
		return nil, "", types.ErrSessionLoginsLocked
	}

	id := "oauth:" + subject
	if session, ok := service.sessions.Get(id); ok {
		if session.State().IsConnected {
			return nil, "", types.ErrSessionAlreadyConnected
		}
		if err := service.sessions.Delete(id); err != nil {
			return nil, "", err
		}
	}

	session, sessionToken, err := service.sessions.Create(id, profile)
	if err == nil {
		service.setExtraData(id, claims, idClaims)
	}

	return session, sessionToken, err
}

func (service *Service) ExtraData(id string) map[string]any {
	service.mu.Lock()
	defer service.mu.Unlock()

	return maps.Clone(service.extra[id])
}

type discovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

func (service *Service) resolveConfig(ctx context.Context) (Config, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	if service.config.IssuerURL == "" || service.issuerResolved {
		return service.config, nil
	}

	issuer, err := url.Parse(service.config.IssuerURL)
	if err != nil || !absoluteURL(service.config.IssuerURL) {
		return Config{}, errors.New("OAuth issuer URL must be an absolute HTTP(S) URL")
	}
	issuer.Path = strings.TrimSuffix(issuer.Path, "/") + "/.well-known/openid-configuration"
	issuer.RawQuery = ""
	issuer.Fragment = ""

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issuer.String(), nil)
	if err != nil {
		return Config{}, err
	}
	req.Header.Set("Accept", "application/json")

	response, err := service.client.Do(req)
	if err != nil {
		return Config{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Config{}, fmt.Errorf("issuer discovery endpoint returned %s", response.Status)
	}

	var metadata discovery
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&metadata); err != nil {
		return Config{}, err
	}

	if metadata.Issuer != service.config.IssuerURL {
		return Config{}, errors.New("issuer discovery response issuer does not match configured issuer")
	}

	for field, value := range map[string]string{
		"authorization_endpoint": metadata.AuthorizationEndpoint,
		"token_endpoint":         metadata.TokenEndpoint,
		"userinfo_endpoint":      metadata.UserInfoEndpoint,
	} {
		if !absoluteURL(value) {
			return Config{}, fmt.Errorf("issuer discovery response has invalid %s", field)
		}
	}

	service.config.AuthorizationURL = metadata.AuthorizationEndpoint
	service.config.TokenURL = metadata.TokenEndpoint
	service.config.UserInfoURL = metadata.UserInfoEndpoint
	service.issuerResolved = true

	return service.config, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

func (service *Service) exchangeCode(ctx context.Context, config Config, code, verifier, redirectURL string) (string, map[string]any, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURL},
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"code_verifier": {verifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	response, err := service.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", nil, fmt.Errorf("token endpoint returned %s", response.Status)
	}

	var payload tokenResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return "", nil, err
	}

	if payload.AccessToken == "" {
		return "", nil, errors.New("token response has no access_token")
	}

	return payload.AccessToken, idTokenClaims(payload.IDToken), nil
}

func (service *Service) userInfo(ctx context.Context, config Config, accessToken string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := service.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("user-info endpoint returned %s", response.Status)
	}

	claims := map[string]any{}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&claims); err != nil {
		return nil, err
	}

	return claims, nil
}

func (service *Service) storeState(value, verifier, redirectURL string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	now := time.Now()
	for key, value := range service.states {
		if now.After(value.expires) {
			delete(service.states, key)
		}
	}

	service.states[value] = state{verifier, redirectURL, now.Add(stateLifetime)}
}

func (service *Service) popState(value string) (state, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()

	result, ok := service.states[value]
	delete(service.states, value)

	return result, ok && time.Now().Before(result.expires)
}

func (service *Service) setExtraData(id string, data ...map[string]any) {
	service.mu.Lock()
	defer service.mu.Unlock()

	totalLen := 0
	for _, d := range data {
		totalLen += len(d)
	}

	result := make(map[string]any, totalLen)
	for _, d := range data {
		maps.Copy(result, d)
	}
	service.extra[id] = result
}

func (service *Service) deleteExtraData(id string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	delete(service.extra, id)
}
