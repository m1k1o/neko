package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

type BroadcastStatusPayload struct {
	URL string `json:"url,omitempty"`
	IsActive bool `json:"is_active"`
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
		utils.HttpBadRequest(w, "Missing broadcast URL.")
		return
	}

	broadcast := h.capture.Broadcast()
	if broadcast.Started() {
		utils.HttpUnprocessableEntity(w, "Server is already broadcasting.")
		return
	}

	if err := broadcast.Start(data.URL); err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	h.sessions.AdminBroadcast(
		message.BroadcastStatus{
			Event:    event.BORADCAST_STATUS,
			IsActive: broadcast.Started(),
			URL:      broadcast.Url(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) boradcastStop(w http.ResponseWriter, r *http.Request) {
	broadcast := h.capture.Broadcast()
	if !broadcast.Started() {
		utils.HttpUnprocessableEntity(w, "Server is not broadcasting.")
		return
	}

	broadcast.Stop()

	h.sessions.AdminBroadcast(
		message.BroadcastStatus{
			Event:    event.BORADCAST_STATUS,
			IsActive: broadcast.Started(),
			URL:      broadcast.Url(),
		}, nil)

	utils.HttpSuccess(w)
}
