package room

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/utils"
)

type ControlStatusPayload struct {
	HasHost bool   `json:"has_host"`
	HostId  string `json:"host_id,omitempty"`
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
			HostId:  host.ID(),
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
	if !session.Profile().CanHost {
		utils.HttpBadRequest(w, "Session is not allowed to host.")
		return
	}

	h.sessions.SetHost(session)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlRelease(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.IsHost() {
		utils.HttpUnprocessableEntity(w, "Session is not the host.")
		return
	}

	if !session.Profile().CanHost {
		utils.HttpBadRequest(w, "Session is not allowed to host.")
		return
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlTake(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.Profile().CanHost {
		utils.HttpBadRequest(w, "Session is not allowed to host.")
		return
	}

	h.sessions.SetHost(session)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlGive(w http.ResponseWriter, r *http.Request) {
	sessionId := chi.URLParam(r, "sessionId")

	target, ok := h.sessions.Get(sessionId)
	if !ok {
		utils.HttpNotFound(w, "Target session was not found.")
		return
	}

	if !target.Profile().CanHost {
		utils.HttpBadRequest(w, "Target session is not allowed to host.")
		return
	}

	h.sessions.SetHost(target)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlReset(w http.ResponseWriter, r *http.Request) {
	host := h.sessions.GetHost()
	if host == nil {
		utils.HttpSuccess(w)
		return
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	utils.HttpSuccess(w)
}
