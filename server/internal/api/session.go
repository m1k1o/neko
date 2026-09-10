package api

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type SessionLoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionDataPayload struct {
	ID      string              `json:"id"`
	Token   string              `json:"token,omitempty"`
	Profile types.MemberProfile `json:"profile"`
	State   types.SessionState  `json:"state"`
}

type AvatarPayload struct {
	Avatar string `json:"avatar"`
}

const maxAvatarDataURLLength = 512 * 1024

func (api *ApiManagerCtx) Login(w http.ResponseWriter, r *http.Request) error {
	data := &SessionLoginPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	session, token, err := api.members.Login(data.Username, data.Password)
	if err != nil {
		if errors.Is(err, types.ErrSessionAlreadyConnected) {
			return utils.HttpUnprocessableEntity("session already connected")
		} else if errors.Is(err, types.ErrMemberDoesNotExist) || errors.Is(err, types.ErrMemberInvalidPassword) {
			return utils.HttpUnauthorized().WithInternalErr(err)
		} else if errors.Is(err, types.ErrSessionLoginsLocked) {
			return utils.HttpForbidden("logins are locked").WithInternalErr(err)
		} else {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}
	}

	sessionData := SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	}

	if api.sessions.CookieEnabled() {
		api.sessions.CookieSetToken(w, token)
	} else {
		sessionData.Token = token
	}

	return utils.HttpSuccess(w, sessionData)
}

func (api *ApiManagerCtx) Logout(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	err := api.members.Logout(session.ID())
	if err != nil {
		if errors.Is(err, types.ErrSessionNotFound) {
			return utils.HttpBadRequest("session is not logged in")
		} else {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}
	}

	if api.sessions.CookieEnabled() {
		api.sessions.CookieClearToken(w, r)
	}

	return utils.HttpSuccess(w, true)
}

func (api *ApiManagerCtx) Whoami(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	return utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	})
}

func (api *ApiManagerCtx) UpdateProfile(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	profile := session.Profile()
	if !profile.IsAdmin {
		// Non-admins may update their display name and avatar only.
		var payload types.MemberProfile
		if err := utils.HttpJsonRequest(w, r, &payload); err != nil {
			return err
		}
		profile.Name = payload.Name
		profile.Avatar = payload.Avatar
	} else {
		if err := utils.HttpJsonRequest(w, r, &profile); err != nil {
			return err
		}
	}

	if err := validateAvatar(profile.Avatar); err != nil {
		return utils.HttpBadRequest(err.Error())
	}

	err := api.members.UpdateProfile(session.ID(), profile)
	if err != nil {
		if errors.Is(err, types.ErrSessionNotFound) {
			return utils.HttpBadRequest("session does not exist")
		} else {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}
	}

	return utils.HttpSuccess(w, true)
}

func (api *ApiManagerCtx) UpdateAvatar(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)
	payload := AvatarPayload{}
	if err := utils.HttpJsonRequest(w, r, &payload); err != nil {
		return err
	}
	if err := validateAvatar(payload.Avatar); err != nil {
		return utils.HttpBadRequest(err.Error())
	}

	profile := session.Profile()
	profile.Avatar = payload.Avatar
	if err := api.members.UpdateProfile(session.ID(), profile); err != nil {
		if errors.Is(err, types.ErrSessionNotFound) {
			return utils.HttpBadRequest("session does not exist")
		}
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w, true)
}

func validateAvatar(value string) error {
	if value == "" {
		return nil
	}
	if len(value) > maxAvatarDataURLLength {
		return fmt.Errorf("avatar is too large; maximum size is %d bytes", maxAvatarDataURLLength)
	}

	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:image/") || !strings.HasSuffix(parts[0], ";base64") {
		return errors.New("avatar must be a base64 image data URL")
	}

	mimeType := strings.TrimPrefix(parts[0], "data:")
	switch mimeType {
	case "image/gif;base64", "image/jpeg;base64", "image/png;base64":
	default:
		return errors.New("avatar format must be PNG, JPEG, or GIF")
	}

	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("avatar contains invalid base64 data")
	}
	if len(raw) == 0 || len(raw) > 384*1024 {
		return errors.New("avatar image is empty or too large")
	}
	if _, _, err := image.DecodeConfig(bytes.NewReader(raw)); err != nil {
		return errors.New("avatar is not a valid image")
	}

	return nil
}

func (api *ApiManagerCtx) Stats(w http.ResponseWriter, r *http.Request) error {
	stats := api.sessions.Stats()
	return utils.HttpSuccess(w, stats)
}
