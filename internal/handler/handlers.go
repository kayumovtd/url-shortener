package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

func PostHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		shortURL, err := svc.Shorten(r.Context(), string(body))
		if err != nil {
			var conflict *service.ErrShortenerConflict
			if errors.As(err, &conflict) {
				utils.WritePlainText(w, http.StatusConflict, conflict.ResultURL)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		utils.WritePlainText(w, http.StatusCreated, shortURL)
	}
}

func GetHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		origURL, err := svc.Unshorten(r.Context(), id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
	}
}

func PingHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Ping(r.Context()); err != nil {
			http.Error(w, "storage not available", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
