package room

import (
	"net/http"

	"github.com/go-chi/render"

	"demodesk/neko/internal/api/utils"
	"demodesk/neko/internal/websocket/broadcast"
)

type ScreenConfiguration struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Rate   int `json:"rate"`
}

func (a *ScreenConfiguration) Bind(r *http.Request) error {
	// Bind will run after the unmarshalling is complete, its a
	// good time to focus some post-processing after a decoding.
	return nil
}

func (a *ScreenConfiguration) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent
	// across the wire
	return nil
}

func (h *RoomHandler) ScreenConfiguration(w http.ResponseWriter, r *http.Request) {
	size := h.remote.GetScreenSize()

	if size == nil {
		render.Render(w, r, utils.ErrMessage(500, "Not implmented."))
		return
	}

	render.JSON(w, r, ScreenConfiguration{
		Width:  size.Width,
		Height: size.Height,
		Rate:   int(size.Rate),
	})
}

func (h *RoomHandler) ScreenConfigurationChange(w http.ResponseWriter, r *http.Request) {
	data := &ScreenConfiguration{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	if err := h.remote.ChangeResolution(data.Width, data.Height, data.Rate); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	if err := broadcast.ScreenConfiguration(h.sessions, "-todo-session-id-", data.Width, data.Height, data.Rate); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	render.JSON(w, r, data)
}

func (h *RoomHandler) ScreenConfigurationsList(w http.ResponseWriter, r *http.Request) {
	list := []render.Renderer{}
	
	ScreenConfigurations := h.remote.ScreenConfigurations()
	for _, size := range ScreenConfigurations {
		for _, fps := range size.Rates {
			list = append(list, &ScreenConfiguration{
				Width:  size.Width,
				Height: size.Height,
				Rate:   int(fps),
			})
		}
	}

	render.RenderList(w, r, list)
}
