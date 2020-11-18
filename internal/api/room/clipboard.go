package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type ClipboardPayload struct {
	Text string `json:"text"`
}

func (h *RoomHandler) ClipboardRead(w http.ResponseWriter, r *http.Request) {
	// TODO: error check?
	text := h.desktop.ReadClipboard()

	utils.HttpSuccess(w, ClipboardPayload{
		Text: text,
	})
}

func (h *RoomHandler) ClipboardWrite(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	// TODO: error check?
	h.desktop.WriteClipboard(data.Text)
	utils.HttpSuccess(w)
}
