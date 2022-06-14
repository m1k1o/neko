package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// Middleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
type metrics struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewMiddleware returns a new prometheus Middleware handler.
func middlewareMetrics(next http.Handler) http.Handler {
	var m metrics

	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "requests_total",
			Namespace: "neko",
			Subsystem: "http",
			Help:      "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.reqs)

	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "request_duration_milliseconds",
		Namespace: "neko",
		Subsystem: "http",
		Help:      "How long it took to process the request, partitioned by status code, method and HTTP path.",
		Buckets:   []float64{300, 1200, 5000},
	},
		[]string{"code", "method", "path"},
	)

	prometheus.MustRegister(m.latency)
	return m.handler(next)
}

func (c metrics) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		c.reqs.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Inc()
		c.latency.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
	}

	return http.HandlerFunc(fn)
}
