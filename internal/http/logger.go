package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type logEntryKey int

const logEntryKeyCtx logEntryKey = iota

func setLogEntry(r *http.Request, data logEntry) *http.Request {
	ctx := context.WithValue(r.Context(), logEntryKeyCtx, data)
	return r.WithContext(ctx)
}

func getLogEntry(r *http.Request) logEntry {
	return r.Context().Value(logEntryKeyCtx).(logEntry)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, setLogEntry(r, newLogEntry(w, r)))
	})
}

type logEntry struct {
	req struct {
		time   time.Time
		id     string
		scheme string
		proto  string
		method string
		remote string
		agent  string
		uri    string
	}
	res struct {
		time  time.Time
		code  int
		bytes int
	}
	err        error
	elapsed    time.Duration
	hasSession bool
	session    types.Session
}

func newLogEntry(w http.ResponseWriter, r *http.Request) logEntry {
	e := logEntry{}
	e.req.time = time.Now()

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		e.req.id = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	e.req.scheme = scheme
	e.req.proto = r.Proto
	e.req.method = r.Method
	e.req.remote = r.RemoteAddr
	e.req.agent = r.UserAgent()
	e.req.uri = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
	return e
}

func (e *logEntry) SetResponse(w http.ResponseWriter, r *http.Request) {
	ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
	e.res.time = time.Now()
	e.res.code = ww.Status()
	e.res.bytes = ww.BytesWritten()

	e.elapsed = e.res.time.Sub(e.req.time)
	e.session, e.hasSession = auth.GetSession(r)
}

func (e *logEntry) SetError(err error) {
	e.err = err
}

func (e *logEntry) Write() {
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
			logger.Err(e.err).Msgf("request failed (%d)", e.res.code)
			return
		}

		if httpErr.Message == "" {
			httpErr.Message = http.StatusText(httpErr.Code)
		}

		logger := logger.Error().Err(httpErr.InternalErr)

		message := httpErr.Message
		if httpErr.InternalMsg != "" {
			message = httpErr.InternalMsg
		}

		logger.Msgf("request failed (%d): %s", e.res.code, message)
		return
	}

	logger.Debug().Msgf("request complete (%d)", e.res.code)
}
