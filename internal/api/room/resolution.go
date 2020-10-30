package room

import (
	"net/http"

	"github.com/go-chi/render"

	"demodesk/neko/internal/api/utils"
)

type ResolutionStruct struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Rate   int `json:"rate"`
}

func (a *ResolutionStruct) Bind(r *http.Request) error {
	return nil
}

func (h *RoomHandler) ResolutionGet(w http.ResponseWriter, r *http.Request) {
	size := h.remote.GetScreenSize()

	if size == nil {
		render.Render(w, r, utils.ErrMessage(500, "Not implmented."))
		return
	}

	render.JSON(w, r, ResolutionStruct{
		Width:  size.Width,
		Height: size.Height,
		Rate:   int(size.Rate),
	})
}

func (h *RoomHandler) ResolutionChange(w http.ResponseWriter, r *http.Request) {
	data := &ResolutionStruct{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	if err := h.remote.ChangeResolution(data.Width, data.Height, data.Rate); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	// TODO: WebSocket notify.

	render.JSON(w, r, data)
}

func (h *RoomHandler) ResolutionList(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, utils.ErrMessage(500, "Not implmented."))
}
