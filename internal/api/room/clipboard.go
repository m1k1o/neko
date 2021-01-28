package room

import (
	"net/http"

	"demodesk/neko/internal/utils"
)

type ClipboardPayload struct {
	Text string `json:"text"`
}

func (h *RoomHandler) clipboardRead(w http.ResponseWriter, r *http.Request) {
	text, err := h.desktop.ReadClipboard()
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w, ClipboardPayload{
		Text: text,
	})
}

func (h *RoomHandler) clipboardWrite(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	err := h.desktop.WriteClipboard(data.Text)
	if err != nil {
		utils.HttpInternalServerError(w, err)
		return
	}

	utils.HttpSuccess(w)
}
