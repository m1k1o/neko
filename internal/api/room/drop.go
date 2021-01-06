package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type DropPayload struct {
	X     int      `json:"x"`
	Y     int      `json:"y"`
	Files []string `json:"files"`
}

func (h *RoomHandler) dropFiles(w http.ResponseWriter, r *http.Request) {
	data := &DropPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	h.desktop.DropFiles(data.X, data.Y, data.Files)
	utils.HttpSuccess(w)
}
