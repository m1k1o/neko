package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type RouterOption func(*router)

func WithRequestID() RouterOption {
	return func(r *router) {
		r.chi.Use(middleware.RequestID)
	}
}

func WithLogger(logger zerolog.Logger) RouterOption {
	return func(r *router) {
		r.chi.Use(middleware.RequestLogger(&logFormatter{logger}))
	}
}

func WithRecoverer() RouterOption {
	return func(r *router) {
		r.chi.Use(middleware.Recoverer)
	}
}

func WithCORS(allowOrigin func(origin string) bool) RouterOption {
	return func(r *router) {
		r.chi.Use(cors.Handler(cors.Options{
			AllowOriginFunc: func(r *http.Request, origin string) bool {
				return allowOrigin(origin)
			},
			AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))
	}
}

func WithPathPrefix(prefix string) RouterOption {
	return func(r *router) {
		r.chi.Use(func(h http.Handler) http.Handler {
			return http.StripPrefix(prefix, h)
		})
	}
}

func WithRealIP() RouterOption {
	return func(r *router) {
		r.chi.Use(middleware.RealIP)
	}
}

type router struct {
	chi chi.Router
}

func newRouter(opts ...RouterOption) types.Router {
	r := &router{chi.NewRouter()}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *router) Group(fn func(types.Router)) {
	r.chi.Group(func(c chi.Router) {
		fn(&router{c})
	})
}

func (r *router) Route(pattern string, fn func(types.Router)) {
	r.chi.Route(pattern, func(c chi.Router) {
		fn(&router{c})
	})
}

func (r *router) Get(pattern string, fn types.RouterHandler) {
	r.chi.Get(pattern, routeHandler(fn))
}

func (r *router) Post(pattern string, fn types.RouterHandler) {
	r.chi.Post(pattern, routeHandler(fn))
}

func (r *router) Put(pattern string, fn types.RouterHandler) {
	r.chi.Put(pattern, routeHandler(fn))
}

func (r *router) Patch(pattern string, fn types.RouterHandler) {
	r.chi.Patch(pattern, routeHandler(fn))
}

func (r *router) Delete(pattern string, fn types.RouterHandler) {
	r.chi.Delete(pattern, routeHandler(fn))
}

func (r *router) With(fn types.MiddlewareHandler) types.Router {
	c := r.chi.With(middlewareHandler(fn))
	return &router{c}
}

func (r *router) Use(fn types.MiddlewareHandler) {
	r.chi.Use(middlewareHandler(fn))
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}

func errorHandler(err error, w http.ResponseWriter, r *http.Request) {
	httpErr, ok := err.(*utils.HTTPError)
	if !ok {
		httpErr = utils.HttpInternalServerError().WithInternalErr(err)
	}

	utils.HttpJsonResponse(w, httpErr.Code, httpErr)
}

func routeHandler(fn types.RouterHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get custom log entry pointer from context
		logEntry, _ := r.Context().Value(middleware.LogEntryCtxKey).(*logEntry)

		if err := fn(w, r); err != nil {
			logEntry.Error(err)
			errorHandler(err, w, r)
		}

		// set session if exits
		if session, ok := auth.GetSession(r); ok {
			logEntry.SetSession(session)
		}
	}
}

func middlewareHandler(fn types.MiddlewareHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get custom log entry pointer from context
			logEntry, _ := r.Context().Value(middleware.LogEntryCtxKey).(*logEntry)

			ctx, err := fn(w, r)
			if err != nil {
				logEntry.Error(err)
				errorHandler(err, w, r)

				// set session if exits
				if session, ok := auth.GetSession(r); ok {
					logEntry.SetSession(session)
				}

				return
			}
			if ctx != nil {
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
