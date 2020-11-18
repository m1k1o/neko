package room

import (
	"net/http"

	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/http/auth"
)

type ScreenConfiguration struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Rate   int `json:"rate"`
}

func (h *RoomHandler) ScreenConfiguration(w http.ResponseWriter, r *http.Request) {
	size := h.desktop.GetScreenSize()

	if size == nil {
		utils.HttpInternalServer(w, "Unable to get screen configuration.")
		return
	}

	utils.HttpSuccess(w, ScreenConfiguration{
		Width:  size.Width,
		Height: size.Height,
		Rate:   int(size.Rate),
	})
}

func (h *RoomHandler) ScreenConfigurationChange(w http.ResponseWriter, r *http.Request) {
	data := &ScreenConfiguration{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if err := h.desktop.ChangeScreenSize(data.Width, data.Height, data.Rate); err != nil {
		utils.HttpUnprocessableEntity(w, err)
		return
	}

	session := auth.GetSession(r)

	h.sessions.Broadcast(
		message.ScreenResolution{
			Event:  event.SCREEN_RESOLUTION,
			ID:     session.ID(),
			Width:  data.Width,
			Height: data.Height,
			Rate:   data.Rate,
		}, nil)

	utils.HttpSuccess(w, data)
}

func (h *RoomHandler) ScreenConfigurationsList(w http.ResponseWriter, r *http.Request) {
	list := []ScreenConfiguration{}
	
	ScreenConfigurations := h.desktop.ScreenConfigurations()
	for _, size := range ScreenConfigurations {
		for _, fps := range size.Rates {
			list = append(list, ScreenConfiguration{
				Width:  size.Width,
				Height: size.Height,
				Rate:   int(fps),
			})
		}
	}

	utils.HttpSuccess(w, list)
}
