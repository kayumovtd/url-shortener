package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

func ShortenHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		shortURL, err := svc.Shorten(r.Context(), req.URL)
		if err != nil {
			var conflict *service.ErrShortenerConflict
			if errors.As(err, &conflict) {
				resp := model.ShortenResponse{Result: conflict.ResultURL}
				utils.WriteJSON(w, http.StatusConflict, resp)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resp := model.ShortenResponse{Result: shortURL}
		utils.WriteJSON(w, http.StatusCreated, resp)
	}
}

func ShortenBatchHandler(svc *service.ShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []model.ShortenBatchRequestItem
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resp, err := svc.ShortenBatch(r.Context(), req)
		if err != nil {
			// Тут может быть как ошибка валидации урлов (bad request),
			// так и ошибка сохранения в стор (internal server error).
			// Можно в будущем добавить более детальную обработку ошибок, пока просто отдаём 400.
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		utils.WriteJSON(w, http.StatusCreated, resp)
	}
}
