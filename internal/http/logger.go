package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

type logFormatter struct {
	logger zerolog.Logger
}

func (l *logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	e := logEntry{logger: l.logger}
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
	err     error
	session *types.Session
	logger  zerolog.Logger
}

func (e *logEntry) SetError(err error) {
	e.err = err
}

func (e *logEntry) SetSession(session types.Session) {
	e.session = &session
}

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	e.res.Time = time.Now()
	e.res.Code = status
	e.res.Bytes = bytes

	logger := e.logger.With().
		Float64("elapsed", float64(elapsed.Nanoseconds())/1000000.0).
		Interface("req", e.req).
		Interface("res", e.res).
		Logger()

	if e.session != nil {
		logger = logger.With().Str("session_id", (*e.session).ID()).Logger()
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

func (e *logEntry) Panic(v interface{}, stack []byte) {
	message := fmt.Sprintf("%+v", v)

	log.Fatal().
		Str("message", message).
		Str("stack", string(stack)).
		Msg("got HTTP panic")
}
