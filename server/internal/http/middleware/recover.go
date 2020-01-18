package middleware

// The original work was derived from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
  "net/http"

  "n.eko.moe/neko/internal/http/endpoint"
)

func Recoverer(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if rvr := recover(); rvr != nil {
        endpoint.WriteError(w, r, rvr)
      }
    }()

    next.ServeHTTP(w, r)
  }

  return http.HandlerFunc(fn)
}
