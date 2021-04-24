package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type ErrResponse struct {
	Message string `json:"message"`
}

func HttpJsonRequest(w http.ResponseWriter, r *http.Request, res interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		if err == io.EOF {
			HttpBadRequest(w, "no data provided")
		} else {
			HttpBadRequest(w, err)
		}

		return false
	}

	return true
}

func HttpJsonResponse(w http.ResponseWriter, status int, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Warn().Err(err).
			Str("module", "http").
			Msg("failed writing json error response")
	}
}

func HttpError(w http.ResponseWriter, status int, res interface{}) {
	HttpJsonResponse(w, status, &ErrResponse{
		Message: fmt.Sprint(res),
	})
}

func HttpSuccess(w http.ResponseWriter, res ...interface{}) {
	if len(res) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		HttpJsonResponse(w, http.StatusOK, res[0])
	}
}

func HttpBadRequest(w http.ResponseWriter, res ...interface{}) {
	defHttpError(w, http.StatusBadRequest, "bad request", res...)
}

func HttpUnauthorized(w http.ResponseWriter, res ...interface{}) {
	defHttpError(w, http.StatusUnauthorized, "invalid or missing access token", res...)
}

func HttpForbidden(w http.ResponseWriter, res ...interface{}) {
	defHttpError(w, http.StatusForbidden, "access token does not have the required scope", res...)
}

func HttpNotFound(w http.ResponseWriter, res ...interface{}) {
	defHttpError(w, http.StatusNotFound, "resource not found", res...)
}

func HttpUnprocessableEntity(w http.ResponseWriter, res ...interface{}) {
	defHttpError(w, http.StatusUnprocessableEntity, "unprocessable entity", res...)
}

func HttpInternalServerError(w http.ResponseWriter, res ...interface{}) {
	defHttpError(w, http.StatusInternalServerError, "internal server error", res...)
}

func defHttpError(w http.ResponseWriter, status int, text string, res ...interface{}) {
	if len(res) == 0 {
		HttpError(w, status, text)
	} else {
		HttpError(w, status, res[0])
	}
}
