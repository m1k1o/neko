package room

import (
	"net/http"
	"strconv"

	"gitlab.com/demodesk/neko/server/pkg/types"
	"gitlab.com/demodesk/neko/server/pkg/types/event"
	"gitlab.com/demodesk/neko/server/pkg/types/message"
	"gitlab.com/demodesk/neko/server/pkg/utils"
)

type ScreenConfigurationPayload struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rate   int16 `json:"rate"`
}

func (h *RoomHandler) screenConfiguration(w http.ResponseWriter, r *http.Request) error {
	size := h.desktop.GetScreenSize()

	if size == nil {
		return utils.HttpInternalServerError().WithInternalMsg("unable to get screen configuration")
	}

	payload := ScreenConfigurationPayload(*size)
	return utils.HttpSuccess(w, payload)
}

func (h *RoomHandler) screenConfigurationChange(w http.ResponseWriter, r *http.Request) error {
	data := &ScreenConfigurationPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	size := types.ScreenSize(*data)
	if err := h.desktop.SetScreenSize(size); err != nil {
		return utils.HttpUnprocessableEntity("cannot set screen size").WithInternalErr(err)
	}

	payload := message.ScreenSize(*data)
	h.sessions.Broadcast(event.SCREEN_UPDATED, payload, nil)

	return utils.HttpSuccess(w, data)
}

func (h *RoomHandler) screenConfigurationsList(w http.ResponseWriter, r *http.Request) error {
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

	return utils.HttpSuccess(w, list)
}

func (h *RoomHandler) screenShotGet(w http.ResponseWriter, r *http.Request) error {
	quality, err := strconv.Atoi(r.URL.Query().Get("quality"))
	if err != nil {
		quality = 90
	}

	img := h.desktop.GetScreenshotImage()
	bytes, err := utils.CreateJPGImage(img, quality)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/jpeg")

	_, err = w.Write(bytes)
	return err
}

func (h *RoomHandler) screenCastGet(w http.ResponseWriter, r *http.Request) error {
	screencast := h.capture.Screencast()
	if !screencast.Enabled() {
		return utils.HttpBadRequest("screencast pipeline is not enabled")
	}

	bytes, err := screencast.Image()
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/jpeg")

	_, err = w.Write(bytes)
	return err
}
