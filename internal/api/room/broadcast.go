package room

import (
	"net/http"

	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
)

type BroadcastStatusPayload struct {
	URL      string `json:"url,omitempty"`
	IsActive bool   `json:"is_active"`
}

func (h *RoomHandler) broadcastStatus(w http.ResponseWriter, r *http.Request) {
	broadcast := h.capture.Broadcast()
	utils.HttpSuccess(w, BroadcastStatusPayload{
		IsActive: broadcast.Started(),
		URL:      broadcast.Url(),
	})
}

func (h *RoomHandler) boradcastStart(w http.ResponseWriter, r *http.Request) {
	data := &BroadcastStatusPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if data.URL == "" {
		utils.HttpBadRequest(w).Msg("missing broadcast URL")
		return
	}

	broadcast := h.capture.Broadcast()
	if broadcast.Started() {
		utils.HttpUnprocessableEntity(w).Msg("server is already broadcasting")
		return
	}

	if err := broadcast.Start(data.URL); err != nil {
		utils.HttpInternalServerError(w, err).Send()
		return
	}

	h.sessions.AdminBroadcast(
		event.BORADCAST_STATUS,
		message.BroadcastStatus{
			IsActive: broadcast.Started(),
			URL:      broadcast.Url(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) boradcastStop(w http.ResponseWriter, r *http.Request) {
	broadcast := h.capture.Broadcast()
	if !broadcast.Started() {
		utils.HttpUnprocessableEntity(w).Msg("server is not broadcasting")
		return
	}

	broadcast.Stop()

	h.sessions.AdminBroadcast(
		event.BORADCAST_STATUS,
		message.BroadcastStatus{
			IsActive: broadcast.Started(),
			URL:      broadcast.Url(),
		}, nil)

	utils.HttpSuccess(w)
}
