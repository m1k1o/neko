package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"
)

func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
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

		entry := &entry{
			fields: fields,
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()

		defer func() {
			entry.Write(ww.Status(), ww.BytesWritten(), time.Since(t1))
		}()

		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}

type entry struct {
	fields map[string]interface{}
	errors []map[string]interface{}
}

func (e *entry) Write(status, bytes int, elapsed time.Duration) {
	res := map[string]interface{}{}
	res["time"] = time.Now().UTC().Format(time.RFC1123)
	res["status"] = status
	res["bytes"] = bytes
	res["elapsed"] = float64(elapsed.Nanoseconds()) / 1000000.0

	e.fields["res"] = res
	e.fields["module"] = "http"

	if len(e.errors) > 0 {
		e.fields["errors"] = e.errors
		log.Error().Fields(e.fields).Msgf("request failed (%d)", status)
	} else {
		log.Debug().Fields(e.fields).Msgf("request complete (%d)", status)
	}
}

func (e *entry) Panic(v interface{}, stack []byte) {
	err := map[string]interface{}{}
	err["message"] = fmt.Sprintf("%+v", v)
	err["stack"] = string(stack)

	e.errors = append(e.errors, err)
}
