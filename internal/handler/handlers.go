package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

func PostHandler(svc *service.ShortenerService, up service.UserProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			utils.WritePlainText(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		userID, ok := RequireUserID(w, r, up)
		if !ok {
			return
		}

		shortURL, err := svc.Shorten(r.Context(), string(body), userID)
		if err != nil {
			var conflict *service.ErrShortenerConflict
			if errors.As(err, &conflict) {
				utils.WritePlainText(w, http.StatusConflict, conflict.ResultURL)
				return
			}
			utils.WritePlainText(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		utils.WritePlainText(w, http.StatusCreated, shortURL)
	}
}

func GetHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		rec, err := svc.Unshorten(r.Context(), id)
		if err != nil {
			utils.WritePlainText(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		if rec.IsDeleted {
			utils.WritePlainText(w, http.StatusGone, http.StatusText(http.StatusGone))
			return
		}
		http.Redirect(w, r, rec.OriginalURL, http.StatusTemporaryRedirect)
	}
}

func PingHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Ping(r.Context()); err != nil {
			utils.WritePlainText(w, http.StatusInternalServerError, "storage not available")
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
