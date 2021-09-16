package room

import (
	"net/http"
	"strconv"

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
		utils.HttpInternalServerError(w, nil).WithInternalMsg("unable to get screen configuration").Send()
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
		Width:  data.Width,
		Height: data.Height,
		Rate:   data.Rate,
	}); err != nil {
		utils.HttpUnprocessableEntity(w).WithInternalErr(err).Msg("cannot set screen size")
		return
	}

	h.sessions.Broadcast(
		event.SCREEN_UPDATED,
		message.ScreenSize{
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

func (h *RoomHandler) screenShotGet(w http.ResponseWriter, r *http.Request) {
	quality, err := strconv.Atoi(r.URL.Query().Get("quality"))
	if err != nil {
		quality = 90
	}

	img := h.desktop.GetScreenshotImage()
	bytes, err := utils.CreateJPGImage(img, quality)
	if err != nil {
		utils.HttpInternalServerError(w, err).Send()
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/jpeg")
	//nolint
	w.Write(bytes)
}

func (h *RoomHandler) screenCastGet(w http.ResponseWriter, r *http.Request) {
	screencast := h.capture.Screencast()
	if !screencast.Enabled() {
		utils.HttpBadRequest(w).Msg("screencast pipeline is not enabled")
		return
	}

	bytes, err := screencast.Image()
	if err != nil {
		utils.HttpInternalServerError(w, err).Send()
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/jpeg")
	//nolint
	w.Write(bytes)
}
