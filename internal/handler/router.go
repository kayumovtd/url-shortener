package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/kayumovtd/url-shortener/internal/middleware"
	"github.com/kayumovtd/url-shortener/internal/repository"
)

func NewRouter(store repository.Store, baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)

	r.Post("/", PostHandler(store, baseURL))
	r.Get("/{id}", GetHandler(store))

	return r
}
