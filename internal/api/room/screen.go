package room

import (
	"net/http"
	"strconv"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
	"github.com/demodesk/neko/pkg/utils"
)

type ScreenConfigurationPayload struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rate   int16 `json:"rate"`
}

func (h *RoomHandler) screenConfiguration(w http.ResponseWriter, r *http.Request) error {
	size := h.desktop.GetScreenSize()

	return utils.HttpSuccess(w, ScreenConfigurationPayload{
		Width:  size.Width,
		Height: size.Height,
		Rate:   size.Rate,
	})
}

func (h *RoomHandler) screenConfigurationChange(w http.ResponseWriter, r *http.Request) error {
	data := &ScreenConfigurationPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	size, err := h.desktop.SetScreenSize(types.ScreenSize{
		Width:  data.Width,
		Height: data.Height,
		Rate:   data.Rate,
	})

	if err != nil {
		return utils.HttpUnprocessableEntity("cannot set screen size").WithInternalErr(err)
	}

	h.sessions.Broadcast(event.SCREEN_UPDATED, message.ScreenSize{
		Width:  size.Width,
		Height: size.Height,
		Rate:   size.Rate,
	})

	return utils.HttpSuccess(w, data)
}

// TODO: remove.
func (h *RoomHandler) screenConfigurationsList(w http.ResponseWriter, r *http.Request) error {
	configurations := h.desktop.ScreenConfigurations()

	list := make([]ScreenConfigurationPayload, 0, len(configurations))
	for _, conf := range configurations {
		list = append(list, ScreenConfigurationPayload{
			Width:  conf.Width,
			Height: conf.Height,
			Rate:   conf.Rate,
		})
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
	// display fallback image when private mode is enabled even if screencast is not
	if session, ok := auth.GetSession(r); ok && session.PrivateModeEnabled() {
		if h.privateModeImage != nil {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Content-Type", "image/jpeg")

			_, err := w.Write(h.privateModeImage)
			return err
		}

		return utils.HttpBadRequest("private mode is enabled but no fallback image available")
	}

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
