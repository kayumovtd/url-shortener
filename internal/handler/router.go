package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/kayumovtd/url-shortener/internal/logger"
	"github.com/kayumovtd/url-shortener/internal/middleware"
	"github.com/kayumovtd/url-shortener/internal/service"
)

func NewRouter(
	svc *service.ShortenerService,
	auth *service.AuthService,
	l *logger.Logger,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.LoggingMiddleware(l))
	r.Use(middleware.AuthMiddleware(auth))

	r.Post("/", PostHandler(svc, auth))
	r.Get("/{id}", GetHandler(svc))
	r.Get("/ping", PingHandler(svc))

	r.Get("/api/user/urls", GetUserURLsHandler(svc, auth))
	r.Post("/api/shorten", ShortenHandler(svc, auth))
	r.Post("/api/shorten/batch", ShortenBatchHandler(svc, auth))
	r.Delete("/api/user/urls", DeleteUserURLsHandler(svc, auth))

	return r
}
