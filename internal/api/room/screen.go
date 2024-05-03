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

func (h *RoomHandler) screenConfiguration(w http.ResponseWriter, r *http.Request) error {
	screenSize := h.desktop.GetScreenSize()

	return utils.HttpSuccess(w, screenSize)
}

func (h *RoomHandler) screenConfigurationChange(w http.ResponseWriter, r *http.Request) error {
	auth, _ := auth.GetSession(r)

	data := &types.ScreenSize{}
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

	h.sessions.Broadcast(event.SCREEN_UPDATED, message.ScreenSizeUpdate{
		ID:         auth.ID(),
		ScreenSize: size,
	})

	return utils.HttpSuccess(w, data)
}

// TODO: remove.
func (h *RoomHandler) screenConfigurationsList(w http.ResponseWriter, r *http.Request) error {
	configurations := h.desktop.ScreenConfigurations()

	return utils.HttpSuccess(w, configurations)
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
