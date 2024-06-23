package room

import (
	"net/http"

	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
	"github.com/demodesk/neko/pkg/utils"
)

type BroadcastStatusPayload struct {
	URL      string `json:"url,omitempty"`
	IsActive bool   `json:"is_active"`
}

func (h *RoomHandler) broadcastStatus(w http.ResponseWriter, r *http.Request) error {
	broadcast := h.capture.Broadcast()

	return utils.HttpSuccess(w, BroadcastStatusPayload{
		IsActive: broadcast.Started(),
		URL:      broadcast.Url(),
	})
}

func (h *RoomHandler) boradcastStart(w http.ResponseWriter, r *http.Request) error {
	data := &BroadcastStatusPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	if data.URL == "" {
		return utils.HttpBadRequest("missing broadcast URL")
	}

	broadcast := h.capture.Broadcast()
	if broadcast.Started() {
		return utils.HttpUnprocessableEntity("server is already broadcasting")
	}

	if err := broadcast.Start(data.URL); err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	h.sessions.AdminBroadcast(
		event.BORADCAST_STATUS,
		message.BroadcastStatus{
			IsActive: broadcast.Started(),
			URL:      broadcast.Url(),
		})

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) boradcastStop(w http.ResponseWriter, r *http.Request) error {
	broadcast := h.capture.Broadcast()
	if !broadcast.Started() {
		return utils.HttpUnprocessableEntity("server is not broadcasting")
	}

	broadcast.Stop()

	h.sessions.AdminBroadcast(
		event.BORADCAST_STATUS,
		message.BroadcastStatus{
			IsActive: broadcast.Started(),
			URL:      broadcast.Url(),
		})

	return utils.HttpSuccess(w)
}
