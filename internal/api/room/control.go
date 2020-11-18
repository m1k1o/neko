package room

import (
	"net/http"

	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/http/auth"
)

type ControlGivePayload struct {
	ID string `json:"id"`
}

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

	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) ControlRelease(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.IsHost() {
		utils.HttpBadRequest(w, "User is not the host.")
		return
	}

	h.sessions.ClearHost()
	
	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) ControlTake(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	h.sessions.SetHost(session)

	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) ControlGive(w http.ResponseWriter, r *http.Request) {
	data := &ControlGivePayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	target, ok := h.sessions.Get(data.ID)
	if !ok {
		utils.HttpBadRequest(w, "Target user was not found.")
		return
	}

	h.sessions.SetHost(target)

	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    target.ID(),
		}, nil)

	utils.HttpSuccess(w)
}
