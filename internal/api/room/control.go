package room

import (
	"net/http"

	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/http/auth"
)

type ControlStatusPayload struct {
	HasHost bool `json:"has_host"`
	HostId string `json:"host_id,omitempty"`
}

type ControlTargetPayload struct {
	ID string `json:"id"`
}

func (h *RoomHandler) controlStatus(w http.ResponseWriter, r *http.Request) {
	host := h.sessions.GetHost()

	if host == nil {
		utils.HttpSuccess(w, ControlStatusPayload{
			HasHost: false,
		})
	} else {
		utils.HttpSuccess(w, ControlStatusPayload{
			HasHost: true,
			HostId: host.ID(),
		})
	}
}

func (h *RoomHandler) controlRequest(w http.ResponseWriter, r *http.Request) {
	host := h.sessions.GetHost()
	if host != nil {
		utils.HttpUnprocessableEntity(w, "There is already a host.")
		return
	}

	session := auth.GetSession(r)
	h.sessions.SetHost(session)

	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlRelease(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.IsHost() {
		utils.HttpUnprocessableEntity(w, "User is not the host.")
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

func (h *RoomHandler) controlTake(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)

	h.sessions.SetHost(session)

	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_LOCKED,
			ID:    session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlGive(w http.ResponseWriter, r *http.Request) {
	data := &ControlTargetPayload{}
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
