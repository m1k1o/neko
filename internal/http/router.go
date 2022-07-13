package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type router struct {
	chi chi.Router
}

func newRouter(logger zerolog.Logger) *router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID) // Create a request ID for each request
	r.Use(middleware.RequestLogger(&logFormatter{logger}))
	r.Use(middleware.Recoverer) // Recover from panics without crashing server
	return &router{r}
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

func (r *router) WithBypass(fn func(next http.Handler) http.Handler) types.Router {
	c := r.chi.With(fn)
	return &router{c}
}

func (r *router) Use(fn types.MiddlewareHandler) {
	r.chi.Use(middlewareHandler(fn))
}

func (r *router) UseBypass(fn func(next http.Handler) http.Handler) {
	r.chi.Use(fn)
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
