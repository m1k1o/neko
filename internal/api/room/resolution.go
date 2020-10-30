package room

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResolutionStruct struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Rate   int `json:"rate"`
}

func (h *RoomHandler) ResolutionGet(w http.ResponseWriter, r *http.Request) {
	size := h.remote.GetScreenSize()

	if size == nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, ErrResponse{
			Code:  -1,
			Message: "Unable to get current screen resolution.",
		})
		return
	}

	render.JSON(w, r, ResolutionStruct{
		Width:  size.Width,
		Height: size.Height,
		Rate:   int(size.Rate),
	})
}

func (h *RoomHandler) ResolutionChange(w http.ResponseWriter, r *http.Request) {
	// data := &ResolutionStruct{}
	// if err := render.Bind(r, data); err != nil {
	// 	render.JSON(w, r, ErrResponse{
	// 		Code:  -1,
	// 		Message: "Invalid Request.",
	// 	})
	// 	return
	// }

	render.JSON(w, r, ErrResponse{
		Code:  -1,
		Message: "Not implmented.",
	})
}

func (h *RoomHandler) ResolutionList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	render.JSON(w, r, ErrResponse{
		Code:  -1,
		Message: "Not implmented.",
	})
}
