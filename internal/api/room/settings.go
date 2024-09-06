package room

import (
	"net/http"

	"github.com/demodesk/neko/pkg/utils"
)

func (h *RoomHandler) settingsGet(w http.ResponseWriter, r *http.Request) error {
	settings := h.sessions.Settings()
	return utils.HttpSuccess(w, settings)
}

func (h *RoomHandler) settingsSet(w http.ResponseWriter, r *http.Request) error {
	settings := h.sessions.Settings()

	if err := utils.HttpJsonRequest(w, r, &settings); err != nil {
		return err
	}

	h.sessions.UpdateSettings(settings)

	return utils.HttpSuccess(w)
}
