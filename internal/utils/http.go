package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func HttpJsonRequest(w http.ResponseWriter, r *http.Request, res interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		if err == io.EOF {
			HttpBadRequest(w).WithInternalErr(err).Msg("no data provided")
		} else {
			HttpBadRequest(w).WithInternalErr(err).Msg("unable to parse provided data")
		}

		return false
	}

	return true
}

func HttpJsonResponse(w http.ResponseWriter, code int, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Err(err).Str("module", "http").Msg("sending http json response failed")
	}
}

func HttpSuccess(w http.ResponseWriter, res ...interface{}) {
	if len(res) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		HttpJsonResponse(w, http.StatusOK, res[0])
	}
}

// HTTPError is an error with a message and an HTTP status code.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`

	InternalErr error  `json:"-"`
	InternalMsg string `json:"-"`

	w http.ResponseWriter `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.InternalMsg != "" {
		return e.InternalMsg
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *HTTPError) Cause() error {
	if e.InternalErr != nil {
		return e.InternalErr
	}
	return e
}

// WithInternalErr adds internal error information to the error
func (e *HTTPError) WithInternalErr(err error) *HTTPError {
	e.InternalErr = err
	return e
}

// WithInternalMsg adds internal message information to the error
func (e *HTTPError) WithInternalMsg(msg string) *HTTPError {
	e.InternalMsg = msg
	return e
}

// WithInternalMsg adds internal formated message information to the error
func (e *HTTPError) WithInternalMsgf(fmtStr string, args ...interface{}) *HTTPError {
	e.InternalMsg = fmt.Sprintf(fmtStr, args...)
	return e
}

// Sends error with custom formated message
func (e *HTTPError) Msgf(fmtSt string, args ...interface{}) {
	e.Message = fmt.Sprintf(fmtSt, args...)
	e.Send()
}

// Sends error with custom message
func (e *HTTPError) Msg(str string) {
	e.Message = str
	e.Send()
}

// Sends error with default status text
func (e *HTTPError) Send() {
	if e.Message == "" {
		e.Message = http.StatusText(e.Code)
	}

	logger := log.Error().
		Err(e.InternalErr).
		Str("module", "http").
		Int("code", e.Code)

	message := e.Message
	if e.InternalMsg != "" {
		message = e.InternalMsg
	}

	logger.Msg(message)
	HttpJsonResponse(e.w, e.Code, e)
}

func HttpError(w http.ResponseWriter, code int) *HTTPError {
	return &HTTPError{
		Code: code,
		w:    w,
	}
}

func HttpBadRequest(w http.ResponseWriter) *HTTPError {
	return HttpError(w, http.StatusBadRequest)
}

func HttpUnauthorized(w http.ResponseWriter) *HTTPError {
	return HttpError(w, http.StatusUnauthorized)
}

func HttpForbidden(w http.ResponseWriter) *HTTPError {
	return HttpError(w, http.StatusForbidden)
}

func HttpNotFound(w http.ResponseWriter) *HTTPError {
	return HttpError(w, http.StatusNotFound)
}

func HttpUnprocessableEntity(w http.ResponseWriter) *HTTPError {
	return HttpError(w, http.StatusUnprocessableEntity)
}

func HttpInternalServerError(w http.ResponseWriter, err error) *HTTPError {
	return HttpError(w, http.StatusInternalServerError).WithInternalErr(err)
}
