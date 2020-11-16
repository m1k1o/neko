package room

import (
	"net/http"

	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/http/auth"
)

func (h *RoomHandler) ControlRequest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if session.IsHost() {
		utils.HttpBadRequest(w, "User is already host.")
		return
	}

	host := h.sessions.GetHost()
	if host != nil {
		utils.HttpBadRequest(w, "There is already a host.")
		return
	}

	h.sessions.SetHost(session)

	if err := h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
	}

	utils.HttpSuccess(w)
}

func (h *RoomHandler) ControlRelease(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.IsHost() {
		utils.HttpBadRequest(w, "User is not the host.")
		return
	}

	h.sessions.ClearHost()

	if err := h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_RELEASE)
	}

	utils.HttpSuccess(w)
}
