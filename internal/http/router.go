package http

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type RouterCtx struct {
	chi chi.Router
}

func newRouter() *RouterCtx {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(LoggerMiddleware)
	return &RouterCtx{router}
}

func (r *RouterCtx) Group(fn func(types.Router)) {
	r.chi.Group(func(c chi.Router) {
		fn(&RouterCtx{c})
	})
}

func (r *RouterCtx) Route(pattern string, fn func(types.Router)) {
	r.chi.Route(pattern, func(c chi.Router) {
		fn(&RouterCtx{c})
	})
}

func (r *RouterCtx) Get(pattern string, fn types.RouterHandler) {
	r.chi.Get(pattern, routeHandler(fn))
}
func (r *RouterCtx) Post(pattern string, fn types.RouterHandler) {
	r.chi.Post(pattern, routeHandler(fn))
}
func (r *RouterCtx) Put(pattern string, fn types.RouterHandler) {
	r.chi.Put(pattern, routeHandler(fn))
}
func (r *RouterCtx) Delete(pattern string, fn types.RouterHandler) {
	r.chi.Delete(pattern, routeHandler(fn))
}

func (r *RouterCtx) With(fn types.MiddlewareHandler) types.Router {
	c := r.chi.With(middlewareHandler(fn))
	return &RouterCtx{c}
}

func (r *RouterCtx) WithBypass(fn func(next http.Handler) http.Handler) types.Router {
	c := r.chi.With(fn)
	return &RouterCtx{c}
}

func (r *RouterCtx) Use(fn types.MiddlewareHandler) {
	r.chi.Use(middlewareHandler(fn))
}
func (r *RouterCtx) UseBypass(fn func(next http.Handler) http.Handler) {
	r.chi.Use(fn)
}

func (r *RouterCtx) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
		logEntry := getLogEntry(r)

		if err := fn(w, r); err != nil {
			logEntry.SetError(err)
			errorHandler(err, w, r)
		}

		logEntry.SetResponse(w, r)
	}
}

func middlewareHandler(fn types.MiddlewareHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logEntry := getLogEntry(r)

			ctx, err := fn(w, r)
			if err != nil {
				logEntry.SetError(err)
				errorHandler(err, w, r)
				logEntry.SetResponse(w, r)
				return
			}
			if ctx != nil {
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
