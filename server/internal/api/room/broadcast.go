package room

import (
	"net/http"

	"github.com/m1k1o/neko/server/pkg/utils"
)

type BroadcastStatusPayload struct {
	URL      string `json:"url,omitempty"`
	IsActive bool   `json:"is_active"`
}

func (h *RoomHandler) broadcastStatus(w http.ResponseWriter, r *http.Request) error {
	active, url := h.desktopApp.BroadcastStatus()

	return utils.HttpSuccess(w, BroadcastStatusPayload{
		IsActive: active,
		URL:      url,
	})
}

func (h *RoomHandler) broadcastStart(w http.ResponseWriter, r *http.Request) error {
	data := &BroadcastStatusPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	if data.URL == "" {
		return utils.HttpBadRequest("missing broadcast URL")
	}

	active, _ := h.desktopApp.BroadcastStatus()
	if active {
		return utils.HttpUnprocessableEntity("server is already broadcasting")
	}

	if err := h.desktopApp.StartBroadcast(data.URL); err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}
	h.desktopApp.BroadcastStatusChanged()

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) broadcastStop(w http.ResponseWriter, r *http.Request) error {
	active, _ := h.desktopApp.BroadcastStatus()
	if !active {
		return utils.HttpUnprocessableEntity("server is not broadcasting")
	}

	h.desktopApp.StopBroadcast()
	h.desktopApp.BroadcastStatusChanged()

	return utils.HttpSuccess(w)
}
