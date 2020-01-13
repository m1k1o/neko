package http

import (
	"net/http"

	"n.eko.moe/neko/internal/http/handler"
)

func New(bind, password, static string) *http.Server {
	return &http.Server{
		Addr:    bind,
		Handler: handler.New(password, static),
	}
}
