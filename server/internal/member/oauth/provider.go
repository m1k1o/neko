package oauth

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
	"strings"
	"sync"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

// MemberProviderCtx disables password authentication. OAuth sessions are
// created by Service after the OAuth callback succeeds.
type MemberProviderCtx struct{}

func New() types.MemberProvider { return &MemberProviderCtx{} }

func (provider *MemberProviderCtx) Connect() error { return nil }

func (provider *MemberProviderCtx) Disconnect() error { return nil }

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	return "", types.MemberProfile{}, types.ErrMemberInvalidPassword
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", errors.New("OAuth members are created by OAuth login")
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return errors.New("OAuth profile is managed by the identity provider")
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return errors.New("OAuth provider does not have passwords")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, errors.New("OAuth members are stored in active sessions")
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return errors.New("OAuth members are managed by the identity provider")
}

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
		service.DeleteExtraData(session.ID())
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
	if !configured(config) {
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
	if !configured(config) {
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
	profile := config.UserProfile
	if administrator(email, boolClaim(claims, "isAdmin") || boolClaim(idClaims, "isAdmin"), config.AdminEmails) {
		profile = config.AdminProfile
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
		service.mu.Lock()
		service.extra[id] = mergeClaims(claims, idClaims)
		service.mu.Unlock()
	}
	return session, sessionToken, err
}

func (service *Service) ExtraData(id string) map[string]any {
	service.mu.Lock()
	defer service.mu.Unlock()
	return cloneClaims(service.extra[id])
}
func (service *Service) DeleteExtraData(id string) {
	service.mu.Lock()
	defer service.mu.Unlock()
	delete(service.extra, id)
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
	for field, value := range map[string]string{"authorization_endpoint": metadata.AuthorizationEndpoint, "token_endpoint": metadata.TokenEndpoint, "userinfo_endpoint": metadata.UserInfoEndpoint} {
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
func configured(config Config) bool {
	return config.Enabled && config.ClientID != "" && config.ClientSecret != "" && config.AuthorizationURL != "" && config.TokenURL != "" && config.UserInfoURL != ""
}
func absoluteURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
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
func (service *Service) exchangeCode(ctx context.Context, config Config, code, verifier, redirectURL string) (string, map[string]any, error) {
	values := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirectURL}, "client_id": {config.ClientID}, "client_secret": {config.ClientSecret}, "code_verifier": {verifier}}
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
	var payload struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
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
func token(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
func challenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
func idTokenClaims(value string) map[string]any {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	claims := map[string]any{}
	if json.Unmarshal(payload, &claims) != nil {
		return nil
	}
	return claims
}
func claim(claims map[string]any, field string) string {
	if value, ok := claims[field]; ok {
		return fmt.Sprint(value)
	}
	return ""
}
func boolClaim(claims map[string]any, field string) bool {
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
	}
	return false
}
func administrator(email string, isAdmin bool, emails []string) bool {
	if isAdmin {
		return true
	}
	for _, candidate := range emails {
		if email != "" && strings.EqualFold(strings.TrimSpace(email), strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}
func mergeClaims(userInfo, idToken map[string]any) map[string]any {
	result := make(map[string]any, len(userInfo)+len(idToken))
	for key, value := range idToken {
		result[key] = value
	}
	for key, value := range userInfo {
		result[key] = value
	}
	return result
}
func cloneClaims(claims map[string]any) map[string]any {
	if len(claims) == 0 {
		return nil
	}
	result := make(map[string]any, len(claims))
	for key, value := range claims {
		result[key] = value
	}
	return result
}
