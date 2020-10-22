package endpoint

import (
  "encoding/json"
  "fmt"
  "net/http"
  "runtime/debug"

  "github.com/go-chi/chi/middleware"
  "github.com/rs/zerolog/log"
)

type (
  Endpoint func(http.ResponseWriter, *http.Request) error

  ErrResponse struct {
    Status    int    `json:"status,omitempty"`
    Err       string `json:"error,omitempty"`
    Message   string `json:"message,omitempty"`
    Details   string `json:"details,omitempty"`
    Code      string `json:"code,omitempty"`
    RequestID string `json:"request,omitempty"`
  }
)

func Handle(handler Endpoint) http.HandlerFunc {
  fn := func(w http.ResponseWriter, r *http.Request) {
    if err := handler(w, r); err != nil {
      WriteError(w, r, err)
    }
  }

  return http.HandlerFunc(fn)
}

var nonErrorsCodes = map[int]bool{
  404: true,
}

func errResponse(input interface{}) *ErrResponse {
  var res *ErrResponse
  var err interface{}

  switch input.(type) {
  case *HandlerError:
    e := input.(*HandlerError)
    res = &ErrResponse{
      Status:  e.Status,
      Err:     http.StatusText(e.Status),
      Message: e.Message,
    }
    err = e.Err
  default:
    res = &ErrResponse{
      Status: http.StatusInternalServerError,
      Err:    http.StatusText(http.StatusInternalServerError),
    }
    err = input
  }

  if err != nil {
    switch err.(type) {
    case *error:
      e := err.(error)
      res.Details = e.Error()
      break
    default:
      res.Details = fmt.Sprintf("%+v", err)
      break
    }
  }

  return res
}

func WriteError(w http.ResponseWriter, r *http.Request, err interface{}) {
  hlog := log.With().
    Str("module", "http").
    Logger()

  res := errResponse(err)

  if reqID := middleware.GetReqID(r.Context()); reqID != "" {
    res.RequestID = reqID
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(res.Status)

  if err := json.NewEncoder(w).Encode(res); err != nil {
    hlog.Warn().Err(err).Msg("Failed writing json error response")
  }

  if !nonErrorsCodes[res.Status] {
    logEntry := middleware.GetLogEntry(r)
    if logEntry != nil {
      logEntry.Panic(err, debug.Stack())
    } else {
      hlog.Error().Str("stack", string(debug.Stack())).Msgf("%+v", err)
    }
  }
}
