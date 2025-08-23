package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kayumovtd/url-shortener/internal/repository"
)

func NewRouter(store repository.Store, baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", PostHandler(store, baseURL))
	r.Get("/{id}", GetHandler(store))

	return r
}
