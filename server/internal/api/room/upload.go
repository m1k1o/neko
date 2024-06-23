package room

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/demodesk/neko/pkg/utils"
)

// TODO: Extract file uploading to custom utility.

// maximum upload size of 32 MB
const maxUploadSize = 32 << 20

func (h *RoomHandler) uploadDrop(w http.ResponseWriter, r *http.Request) error {
	if !h.desktop.IsUploadDropEnabled() {
		return utils.HttpBadRequest("upload drop is disabled")
	}

	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return utils.HttpBadRequest("failed to parse multipart form").WithInternalErr(err)
	}

	//nolint
	defer r.MultipartForm.RemoveAll()

	X, err := strconv.Atoi(r.FormValue("x"))
	if err != nil {
		return utils.HttpBadRequest("no X coordinate received").WithInternalErr(err)
	}

	Y, err := strconv.Atoi(r.FormValue("y"))
	if err != nil {
		return utils.HttpBadRequest("no Y coordinate received").WithInternalErr(err)
	}

	req_files := r.MultipartForm.File["files"]
	if len(req_files) == 0 {
		return utils.HttpBadRequest("no files received")
	}

	dir, err := os.MkdirTemp("", "neko-drop-*")
	if err != nil {
		return utils.HttpInternalServerError().
			WithInternalErr(err).
			WithInternalMsg("unable to create temporary directory")
	}

	files := []string{}
	for _, req_file := range req_files {
		path := path.Join(dir, req_file.Filename)

		srcFile, err := req_file.Open()
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to open uploaded file")
		}

		defer srcFile.Close()

		dstFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to open destination file")
		}

		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to copy uploaded file to destination file")
		}

		files = append(files, path)
	}

	if !h.desktop.DropFiles(X, Y, files) {
		return utils.HttpInternalServerError().
			WithInternalMsg("unable to drop files")
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) uploadDialogPost(w http.ResponseWriter, r *http.Request) error {
	if !h.desktop.IsFileChooserDialogEnabled() {
		return utils.HttpBadRequest("file chooser dialog is disabled")
	}

	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return utils.HttpBadRequest("failed to parse multipart form").WithInternalErr(err)
	}

	//nolint
	defer r.MultipartForm.RemoveAll()

	req_files := r.MultipartForm.File["files"]
	if len(req_files) == 0 {
		return utils.HttpBadRequest("no files received")
	}

	if !h.desktop.IsFileChooserDialogOpened() {
		return utils.HttpUnprocessableEntity("file chooser dialog is not open")
	}

	dir, err := os.MkdirTemp("", "neko-dialog-*")
	if err != nil {
		return utils.HttpInternalServerError().
			WithInternalErr(err).
			WithInternalMsg("unable to create temporary directory")
	}

	for _, req_file := range req_files {
		path := path.Join(dir, req_file.Filename)

		srcFile, err := req_file.Open()
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to open uploaded file")
		}

		defer srcFile.Close()

		dstFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to open destination file")
		}

		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return utils.HttpInternalServerError().
				WithInternalErr(err).
				WithInternalMsg("unable to copy uploaded file to destination file")
		}
	}

	if err := h.desktop.HandleFileChooserDialog(dir); err != nil {
		return utils.HttpInternalServerError().
			WithInternalErr(err).
			WithInternalMsg("unable to handle file chooser dialog")
	}

	return utils.HttpSuccess(w)
}

func (h *RoomHandler) uploadDialogClose(w http.ResponseWriter, r *http.Request) error {
	if !h.desktop.IsFileChooserDialogEnabled() {
		return utils.HttpBadRequest("file chooser dialog is disabled")
	}

	if !h.desktop.IsFileChooserDialogOpened() {
		return utils.HttpUnprocessableEntity("file chooser dialog is not open")
	}

	h.desktop.CloseFileChooserDialog()

	return utils.HttpSuccess(w)
}
