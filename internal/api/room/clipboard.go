package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type ClipboardData struct {
	Text string `json:"text"`
}

func (h *RoomHandler) ClipboardRead(w http.ResponseWriter, r *http.Request) {
	// TODO: error check?
	text := h.desktop.ReadClipboard()

	utils.HttpSuccess(w, ClipboardData{
		Text: text,
	})
}

func (h *RoomHandler) ClipboardWrite(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardData{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	// TODO: error check?
	h.desktop.WriteClipboard(data.Text)
	utils.HttpSuccess(w)
}
