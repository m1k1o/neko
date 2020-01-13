package handler

import (
	"net/http"
)

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) error {
	return h.manager.Upgrade(w, r)
}
