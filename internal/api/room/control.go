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
		utils.HttpUnprocessableEntity(w).Msg("there is already a host")
		return
	}

	session := auth.GetSession(r)
	h.sessions.SetHost(session)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlRelease(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.IsHost() {
		utils.HttpUnprocessableEntity(w).Msg("session is not the host")
		return
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlTake(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	h.sessions.SetHost(session)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlGive(w http.ResponseWriter, r *http.Request) {
	sessionId := chi.URLParam(r, "sessionId")

	target, ok := h.sessions.Get(sessionId)
	if !ok {
		utils.HttpNotFound(w).Msg("target session was not found")
		return
	}

	if !target.Profile().CanHost {
		utils.HttpBadRequest(w).Msg("target session is not allowed to host")
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
