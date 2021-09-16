package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type logEntryKey int

const logEntryKeyCtx logEntryKey = iota

func setLogEntry(r *http.Request, data *logEntry) *http.Request {
	ctx := context.WithValue(r.Context(), logEntryKeyCtx, data)
	return r.WithContext(ctx)
}

func getLogEntry(r *http.Request) *logEntry {
	return r.Context().Value(logEntryKeyCtx).(*logEntry)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		logEntry := newLogEntry(w, r)
		defer func() {
			logEntry.Write(ww.Status(), ww.BytesWritten())
		}()

		next.ServeHTTP(ww, setLogEntry(r, logEntry))
	})
}

type logEntry struct {
	req struct {
		Time   time.Time
		Id     string
		Scheme string
		Proto  string
		Method string
		Remote string
		Agent  string
		Uri    string
	}
	res struct {
		Time  time.Time
		Code  int
		Bytes int
	}
	err        error
	elapsed    time.Duration
	hasSession bool
	session    types.Session
}

func newLogEntry(w http.ResponseWriter, r *http.Request) *logEntry {
	e := logEntry{}
	e.req.Time = time.Now()

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		e.req.Id = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	e.req.Scheme = scheme
	e.req.Proto = r.Proto
	e.req.Method = r.Method
	e.req.Remote = r.RemoteAddr
	e.req.Agent = r.UserAgent()
	e.req.Uri = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
	return &e
}

func (e *logEntry) SetResponse(w http.ResponseWriter, r *http.Request) {
	e.res.Time = time.Now()

	e.elapsed = e.res.Time.Sub(e.req.Time)
	e.session, e.hasSession = auth.GetSession(r)
}

func (e *logEntry) SetError(err error) {
	e.err = err
}

func (e *logEntry) Write(status, bytes int) {
	e.res.Code = status
	e.res.Bytes = bytes

	logger := log.With().
		Str("module", "http").
		Float64("elapsed", float64(e.elapsed.Nanoseconds())/1000000.0).
		Interface("req", e.req).
		Interface("res", e.res).
		Logger()

	if e.hasSession {
		logger = logger.With().Str("session_id", e.session.ID()).Logger()
	}

	if e.err != nil {
		httpErr, ok := e.err.(*utils.HTTPError)
		if !ok {
			logger.Err(e.err).Msgf("request failed (%d)", e.res.Code)
			return
		}

		if httpErr.Message == "" {
			httpErr.Message = http.StatusText(httpErr.Code)
		}

		var logEvent *zerolog.Event
		if httpErr.Code < 500 {
			logEvent = logger.Warn()
		} else {
			logEvent = logger.Error()
		}

		message := httpErr.Message
		if httpErr.InternalMsg != "" {
			message = httpErr.InternalMsg
		}

		logEvent.Err(httpErr.InternalErr).Msgf("request failed (%d): %s", e.res.Code, message)
		return
	}

	logger.Debug().Msgf("request complete (%d)", e.res.Code)
}
