package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type logformatter struct {
	logger zerolog.Logger
}

func (l *logformatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	req := map[string]interface{}{}

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

	fields := map[string]interface{}{}
	fields["req"] = req

	return &logentry{
		fields: fields,
		logger: l.logger,
	}
}

type logentry struct {
	logger zerolog.Logger
	fields map[string]interface{}
	errors []map[string]interface{}
}

func (e *logentry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	res := map[string]interface{}{}
	res["time"] = time.Now().UTC().Format(time.RFC1123)
	res["status"] = status
	res["bytes"] = bytes
	res["elapsed"] = float64(elapsed.Nanoseconds()) / 1000000.0

	e.fields["res"] = res
	e.fields["module"] = "http"

	if len(e.errors) > 0 {
		e.fields["errors"] = e.errors
		e.logger.Error().Fields(e.fields).Msgf("request failed (%d)", status)
	} else {
		e.logger.Debug().Fields(e.fields).Msgf("request complete (%d)", status)
	}
}

func (e *logentry) Panic(v interface{}, stack []byte) {
	err := map[string]interface{}{}
	err["message"] = fmt.Sprintf("%+v", v)
	err["stack"] = string(stack)

	e.errors = append(e.errors, err)
}
