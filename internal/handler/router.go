package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/kayumovtd/url-shortener/internal/middleware"
	"github.com/kayumovtd/url-shortener/internal/service"
)

func NewRouter(svc *service.ShortenerService) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.LoggingMiddleware)

	r.Post("/", PostHandler(svc))
	r.Get("/{id}", GetHandler(svc))
	r.Post("/api/shorten", ShortenHandler(svc))

	return r
}
