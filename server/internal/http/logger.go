package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type logFormatter struct {
	logger zerolog.Logger
}

func (l *logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	// exclude health & metrics from logs
	if r.RequestURI == "/health" || r.RequestURI == "/metrics" {
		return &nulllog{}
	}

	req := map[string]any{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		req["id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	req["scheme"] = scheme
	req["proto"] = r.Proto
	req["method"] = r.Method
	req["remote"] = r.RemoteAddr
	req["agent"] = r.UserAgent()
	req["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	return &logEntry{
		logger: l.logger.With().Interface("req", req).Logger(),
	}
}

type logEntry struct {
	logger  zerolog.Logger
	err     error
	panic   *logPanic
	session types.Session
}

type logPanic struct {
	message string
	stack   string
}

func (e *logEntry) Panic(v any, stack []byte) {
	e.panic = &logPanic{
		message: fmt.Sprintf("%+v", v),
		stack:   string(stack),
	}
}

func (e *logEntry) Error(err error) {
	e.err = err
}

func (e *logEntry) SetSession(session types.Session) {
	e.session = session
}

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
	res := map[string]any{}
	res["time"] = time.Now().UTC().Format(time.RFC1123)
	res["status"] = status
	res["bytes"] = bytes
	res["elapsed"] = float64(elapsed.Nanoseconds()) / 1000000.0

	logger := e.logger.With().Interface("res", res).Logger()

	// add session ID to logs (if exists)
	if e.session != nil {
		logger = logger.With().Str("session_id", e.session.ID()).Logger()
	}

	// handle panic error message
	if e.panic != nil {
		logger.WithLevel(zerolog.PanicLevel).
			Err(e.err).
			Str("stack", e.panic.stack).
			Msgf("request failed (%d): %s", status, e.panic.message)
		return
	}

	// handle panic error message
	if e.err != nil {
		httpErr, ok := e.err.(*utils.HTTPError)
		if !ok {
			logger.Err(e.err).Msgf("request failed (%d)", status)
			return
		}

		if httpErr.Message == "" {
			httpErr.Message = http.StatusText(httpErr.Code)
		}

		var logLevel zerolog.Level
		if httpErr.Code < 500 {
			logLevel = zerolog.WarnLevel
		} else {
			logLevel = zerolog.ErrorLevel
		}

		message := httpErr.Message
		if httpErr.InternalMsg != "" {
			message = httpErr.InternalMsg
		}

		logger.WithLevel(logLevel).Err(httpErr.InternalErr).Msgf("request failed (%d): %s", status, message)
		return
	}

	logger.Debug().Msgf("request complete (%d)", status)
}

type nulllog struct{}

func (e *nulllog) Panic(v any, stack []byte)        {}
func (e *nulllog) Error(err error)                  {}
func (e *nulllog) SetSession(session types.Session) {}
func (e *nulllog) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
}
