package room

import (
	// TODO: Unused now.
	//"bytes"
	//"strings"

	"net/http"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type ClipboardPayload struct {
	Text string `json:"text,omitempty"`
	HTML string `json:"html,omitempty"`
}

func (h *RoomHandler) clipboardGetText(w http.ResponseWriter, r *http.Request) error {
	data, err := h.desktop.ClipboardGetText()
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w, ClipboardPayload{
		Text: data.Text,
		HTML: data.HTML,
	})
}

func (h *RoomHandler) clipboardSetText(w http.ResponseWriter, r *http.Request) error {
	data := &ClipboardPayload{}
	if err := utils.HttpJsonRequest(w, r, data); err != nil {
		return err
	}

	err := h.desktop.ClipboardSetText(types.ClipboardText{
		Text: data.Text,
		HTML: data.HTML,
	})

	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) clipboardGetImage(w http.ResponseWriter, r *http.Request) error {
	bytes, err := h.desktop.ClipboardGetBinary("image/png")
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/png")

	_, err = w.Write(bytes)
	return err
}

/* TODO: Unused now.
func (h *RoomHandler) clipboardSetImage(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseMultipartForm(MAX_UPLOAD_SIZE)
	if err != nil {
		return utils.HttpBadRequest("failed to parse multipart form").WithInternalErr(err)
	}

	//nolint
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		return utils.HttpBadRequest("no file received").WithInternalErr(err)
	}

	defer file.Close()

	mime := header.Header.Get("Content-Type")
	if !strings.HasPrefix(mime, "image/") {
		return utils.HttpBadRequest("file must be image")
	}

	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(file)
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err).WithInternalMsg("unable to read from uploaded file")
	}

	err = h.desktop.ClipboardSetBinary("image/png", buffer.Bytes())
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err).WithInternalMsg("unable set image to clipboard")
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) clipboardGetTargets(w http.ResponseWriter, r *http.Request) error {
	targets, err := h.desktop.ClipboardGetTargets()
	if err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err)
	}

	return utils.HttpSuccess(w, targets)
}

*/
