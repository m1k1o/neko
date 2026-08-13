package api

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/m1k1o/neko/server/internal/member/oauth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

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
	return &oauthHandler{service: service, pathPrefix: pathPrefix, trustProxy: trustProxy, providerOAuth: providerOAuth}
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
	http.Redirect(w, r, location, http.StatusFound)
	return nil
}

func (api *ApiManagerCtx) OAuthCallback(w http.ResponseWriter, r *http.Request) error {
	if !api.oauth.providerOAuth || !api.oauth.service.Enabled() {
		return utils.HttpNotFound()
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		return utils.HttpUnauthorized("OAuth authorization was declined").WithInternalMsg(providerError)
	}
	_, token, err := api.oauth.service.Complete(r.Context(), r.URL.Query().Get("state"), r.URL.Query().Get("code"))
	if err != nil {
		return oauthServiceError(err, false)
	}
	api.sessions.CookieSetToken(w, token)
	http.Redirect(w, r, api.oauth.successRedirect(), http.StatusSeeOther)
	return nil
}

func oauthServiceError(err error, start bool) error {
	if errors.Is(err, types.ErrSessionAlreadyConnected) {
		return utils.HttpUnprocessableEntity("session already connected")
	}
	if errors.Is(err, types.ErrSessionLoginsLocked) {
		return utils.HttpForbidden("logins are locked").WithInternalErr(err)
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
	callback := (&url.URL{Scheme: scheme, Host: host, Path: path.Join(handler.pathPrefix, "/api/oauth/callback")}).String()
	if parsed, err := url.Parse(callback); err != nil || parsed.Host == "" {
		return "", errors.New("request has no valid public URL")
	}
	return callback, nil
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
