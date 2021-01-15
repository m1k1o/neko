package room

import (
	"net/http"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
)

type ScreenConfigurationPayload struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rate   int16 `json:"rate"`
}

func (h *RoomHandler) screenConfiguration(w http.ResponseWriter, r *http.Request) {
	size := h.desktop.GetScreenSize()

	if size == nil {
		utils.HttpInternalServerError(w, "Unable to get screen configuration.")
		return
	}

	utils.HttpSuccess(w, ScreenConfigurationPayload{
		Width:  size.Width,
		Height: size.Height,
		Rate:   size.Rate,
	})
}

func (h *RoomHandler) screenConfigurationChange(w http.ResponseWriter, r *http.Request) {
	data := &ScreenConfigurationPayload{}
	if !utils.HttpJsonRequest(w, r, data) {
		return
	}

	if err := h.desktop.SetScreenSize(types.ScreenSize{
		Width: data.Width,
		Height: data.Height,
		Rate: data.Rate,
	}); err != nil {
		utils.HttpUnprocessableEntity(w, err)
		return
	}

	h.sessions.Broadcast(
		message.ScreenSize{
			Event:  event.SCREEN_UPDATED,
			Width:  data.Width,
			Height: data.Height,
			Rate:   data.Rate,
		}, nil)

	utils.HttpSuccess(w, data)
}

func (h *RoomHandler) screenConfigurationsList(w http.ResponseWriter, r *http.Request) {
	list := []ScreenConfigurationPayload{}
	
	ScreenConfigurations := h.desktop.ScreenConfigurations()
	for _, size := range ScreenConfigurations {
		for _, fps := range size.Rates {
			list = append(list, ScreenConfigurationPayload{
				Width:  size.Width,
				Height: size.Height,
				Rate:   fps,
			})
		}
	}

	utils.HttpSuccess(w, list)
}
