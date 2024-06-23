package room

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

func (h *RoomHandler) settingsGet(w http.ResponseWriter, r *http.Request) error {
	settings := h.sessions.Settings()
	return utils.HttpSuccess(w, settings)
}

func (h *RoomHandler) settingsSet(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	// We read the request body first and unmashal it inside the UpdateSettingsFunc
	// to ensure atomicity of the operation.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return utils.HttpBadRequest("unable to read request body").WithInternalErr(err)
	}

	h.sessions.UpdateSettingsFunc(session, func(settings *types.Settings) bool {
		err = json.Unmarshal(body, settings)
		return err == nil
	})

	if err != nil {
		return utils.HttpBadRequest("unable to parse provided data").WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}
