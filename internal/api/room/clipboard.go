package room

import (
	"net/http"

	"github.com/go-chi/render"

	"demodesk/neko/internal/api/utils"
)

type ClipboardData struct {
	Text string `json:"text"`
}

func (a *ClipboardData) Bind(r *http.Request) error {
	// Bind will run after the unmarshalling is complete, its a
	// good time to focus some post-processing after a decoding.
	return nil
}

func (a *ClipboardData) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent
	// across the wire
	return nil
}

func (h *RoomHandler) ClipboardRead(w http.ResponseWriter, r *http.Request) {
	// TODO: error check?
	text := h.desktop.ReadClipboard()

	render.JSON(w, r, ClipboardData{
		Text: text,
	})
}

func (h *RoomHandler) ClipboardWrite(w http.ResponseWriter, r *http.Request) {
	data := &ClipboardData{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, utils.ErrBadRequest(err))
		return
	}

	// TODO: error check?
	h.desktop.WriteClipboard(data.Text)

	w.WriteHeader(http.StatusNoContent)
}
