package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

func PostHandler(store repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		originalURL := strings.TrimSpace(string(body))
		u, err := url.ParseRequestURI(originalURL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		originalURL = u.String()

		shortID := utils.GenerateID(originalURL)
		shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortID)

		if err := store.Set(shortID, originalURL); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func GetHandler(store repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		origURL, ok := store.Get(id)
		if !ok {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
	}
}

func mainRouter(store repository.Store) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", PostHandler(store))
	r.Get("/{id}", GetHandler(store))

	return r
}

func main() {
	store := repository.NewInMemoryStore()
	log.Fatal(http.ListenAndServe(":8080", mainRouter(store)))
}
