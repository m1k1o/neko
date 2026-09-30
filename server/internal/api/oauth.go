package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/m1k1o/neko/server/internal/member/oauth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

const oauthStateCookiePrefix = "NEKO_OAUTH_STATE_"

type OAuthUIConfig struct {
	Enabled              bool   `json:"enabled"`
	Name                 string `json:"name"`
	LoginURL             string `json:"login_url"`
	PasswordLoginEnabled bool   `json:"password_login_enabled"`
}

// oauthHandler is deliberately limited to HTTP concerns. The OAuth protocol,
// profile mapping, and session creation are owned by member/oauth.Service.
type oauthHandler struct {
	service       *oauth.Service
	pathPrefix    string
	trustProxy    bool
	providerOAuth bool
}

func newOAuthHandler(service *oauth.Service, pathPrefix string, trustProxy, providerOAuth bool) *oauthHandler {
	if pathPrefix == "" {
		pathPrefix = "/"
	}

	return &oauthHandler{
		service:       service,
		pathPrefix:    pathPrefix,
		trustProxy:    trustProxy,
		providerOAuth: providerOAuth,
	}
}

func (api *ApiManagerCtx) OAuthConfig(w http.ResponseWriter, r *http.Request) error {
	return utils.HttpSuccess(w, OAuthUIConfig{
		Enabled:              api.oauth.providerOAuth && api.oauth.service.Enabled(),
		Name:                 api.oauth.service.Name(),
		LoginURL:             path.Join(api.oauth.pathPrefix, "/api/oauth/login"),
		PasswordLoginEnabled: !api.oauth.providerOAuth,
	})
}

func (api *ApiManagerCtx) OAuthLogin(w http.ResponseWriter, r *http.Request) error {
	if !api.oauth.providerOAuth || !api.oauth.service.Enabled() {
		return utils.HttpNotFound()
	}

	if !api.sessions.CookieEnabled() {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth requires session cookies to be enabled")
	}

	callbackURL, err := api.oauth.callbackURL(r)
	if err != nil {
		return utils.HttpInternalServerError("unable to determine OAuth callback URL").WithInternalErr(err)
	}

	location, err := api.oauth.service.Start(r.Context(), callbackURL)
	if err != nil {
		return oauthServiceError(err, true)
	}
	if err := api.oauth.setStateCookie(w, location, callbackURL); err != nil {
		return utils.HttpInternalServerError("unable to initialize OAuth state").WithInternalErr(err)
	}

	http.Redirect(w, r, location, http.StatusFound)
	return nil
}

func (api *ApiManagerCtx) OAuthCallback(w http.ResponseWriter, r *http.Request) error {
	if !api.oauth.providerOAuth || !api.oauth.service.Enabled() {
		return utils.HttpNotFound()
	}

	stateToken := r.URL.Query().Get("state")
	if !api.oauth.stateCookieMatches(r, stateToken) {
		return utils.HttpBadRequest("OAuth state is invalid or expired")
	}
	api.oauth.clearStateCookie(w, stateToken)

	if providerError := r.URL.Query().Get("error"); providerError != "" {
		return utils.HttpUnauthorized("OAuth authorization was declined").WithInternalMsg(providerError)
	}

	_, token, err := api.oauth.service.Complete(r.Context(), stateToken, r.URL.Query().Get("code"))
	if err != nil {
		return oauthServiceError(err, false)
	}
	api.sessions.CookieSetToken(w, token)

	http.Redirect(w, r, api.oauth.successRedirect(), http.StatusSeeOther)
	return nil
}

func (handler *oauthHandler) setStateCookie(w http.ResponseWriter, location, callbackURL string) error {
	authorizationURL, err := url.Parse(location)
	if err != nil {
		return err
	}

	stateToken := authorizationURL.Query().Get("state")
	redirectURL, err := url.Parse(authorizationURL.Query().Get("redirect_uri"))
	if stateToken == "" || err != nil || !httpURL(redirectURL) {
		return errors.New("OAuth authorization URL is missing a valid state or redirect_uri")
	}

	requestCallbackURL, err := url.Parse(callbackURL)
	if err != nil || !sameCookieHost(redirectURL, requestCallbackURL) {
		return errors.New("OAuth redirect_uri host does not match the login request host")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName(stateToken),
		Value:    stateToken,
		Path:     "/",
		MaxAge:   int(oauth.StateLifetime / time.Second),
		Expires:  time.Now().Add(oauth.StateLifetime),
		Secure:   redirectURL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (handler *oauthHandler) stateCookieMatches(r *http.Request, stateToken string) bool {
	if stateToken == "" {
		return false
	}

	cookie, err := r.Cookie(oauthStateCookieName(stateToken))
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(stateToken)) == 1
}

func (handler *oauthHandler) clearStateCookie(w http.ResponseWriter, stateToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName(stateToken),
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func oauthStateCookieName(stateToken string) string {
	digest := sha256.Sum256([]byte(stateToken))
	return oauthStateCookiePrefix + hex.EncodeToString(digest[:16])
}

func httpURL(value *url.URL) bool {
	return value != nil && value.Host != "" && (value.Scheme == "http" || value.Scheme == "https")
}

func sameCookieHost(left, right *url.URL) bool {
	return httpURL(left) && httpURL(right) && strings.EqualFold(left.Hostname(), right.Hostname())
}

func oauthServiceError(err error, start bool) error {
	if errors.Is(err, types.ErrSessionAlreadyConnected) {
		return utils.HttpUnprocessableEntity("session already connected")
	}

	if errors.Is(err, types.ErrSessionLoginsLocked) {
		return utils.HttpForbidden("logins are locked").WithInternalErr(err)
	}

	if errors.Is(err, types.ErrSessionEmailNotAllowed) {
		return utils.HttpForbidden("email address is not allowed to log in").WithInternalErr(err)
	}

	if start && strings.HasPrefix(err.Error(), "issuer discovery:") {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth issuer discovery failed").WithInternalErr(err)
	}

	if start {
		return utils.HttpError(http.StatusServiceUnavailable, err.Error()).WithInternalErr(err)
	}

	if strings.Contains(err.Error(), "missing state or code") || strings.Contains(err.Error(), "state is invalid") {
		return utils.HttpBadRequest(err.Error()).WithInternalErr(err)
	}

	if strings.HasPrefix(err.Error(), "issuer discovery:") || strings.Contains(err.Error(), "not fully configured") {
		return utils.HttpError(http.StatusServiceUnavailable, "OAuth issuer discovery failed").WithInternalErr(err)
	}

	return utils.HttpUnauthorized("OAuth authorization failed").WithInternalErr(err)
}

func (handler *oauthHandler) callbackURL(r *http.Request) (string, error) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host
	if handler.trustProxy {
		if value := forwardedHeaderValue(r.Header.Get("X-Forwarded-Proto")); value != "" {
			scheme = value
		}
		if value := forwardedHeaderValue(r.Header.Get("X-Forwarded-Host")); value != "" {
			host = value
		}
	}

	if host == "" || (scheme != "http" && scheme != "https") {
		return "", errors.New("request has no valid public URL")
	}

	callback := (&url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   handler.callbackPath(),
	}).String()
	if parsed, err := url.Parse(callback); err != nil || parsed.Host == "" {
		return "", errors.New("request has no valid public URL")
	}

	return callback, nil
}

func (handler *oauthHandler) callbackPath() string {
	return path.Join(handler.pathPrefix, "/api/oauth/callback")
}

func (handler *oauthHandler) successRedirect() string {
	redirect := handler.service.SuccessRedirect()
	if strings.HasPrefix(redirect, "/") && !strings.HasPrefix(redirect, "//") {
		return path.Join(handler.pathPrefix, redirect)
	}

	return handler.pathPrefix
}

func forwardedHeaderValue(value string) string {
	return strings.TrimSpace(strings.Split(value, ",")[0])
}
