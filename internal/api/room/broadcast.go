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
	utils.HttpSuccess(w, BroadcastStatusPayload{
		IsActive: h.capture.BroadcastEnabled(),
		URL:      h.capture.BroadcastUrl(),
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

	if h.capture.BroadcastEnabled() {
		utils.HttpBadRequest(w, "Server is already broadcasting.")
		return
	}

	if err := h.capture.StartBroadcast(data.URL); err != nil {
		utils.HttpInternalServer(w, err)
		return
	}

	h.sessions.AdminBroadcast(
		message.BroadcastStatus{
			Event:    event.BORADCAST_STATUS,
			IsActive: h.capture.BroadcastEnabled(),
			URL:      h.capture.BroadcastUrl(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) boradcastStop(w http.ResponseWriter, r *http.Request) {
	if !h.capture.BroadcastEnabled() {
		utils.HttpBadRequest(w, "Server is not broadcasting.")
		return
	}
	
	h.capture.StopBroadcast()

	h.sessions.AdminBroadcast(
		message.BroadcastStatus{
			Event:    event.BORADCAST_STATUS,
			IsActive: h.capture.BroadcastEnabled(),
			URL:      h.capture.BroadcastUrl(),
		}, nil)

	utils.HttpSuccess(w)
}
