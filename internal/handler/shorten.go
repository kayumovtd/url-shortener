package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

func ShortenHandler(svc *service.ShortenerService, up service.UserProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		userID, ok := RequireUserID(w, r, up)
		if !ok {
			return
		}

		shortURL, err := svc.Shorten(r.Context(), req.URL, userID)
		if err != nil {
			var conflict *service.ErrShortenerConflict
			if errors.As(err, &conflict) {
				resp := model.ShortenResponse{Result: conflict.ResultURL}
				utils.WriteJSON(w, http.StatusConflict, resp)
				return
			}
			utils.WriteJSONError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		resp := model.ShortenResponse{Result: shortURL}
		utils.WriteJSON(w, http.StatusCreated, resp)
	}
}

func ShortenBatchHandler(svc *service.ShortenerService, up service.UserProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []model.ShortenBatchRequestItem
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		userID, ok := RequireUserID(w, r, up)
		if !ok {
			return
		}

		resp, err := svc.ShortenBatch(r.Context(), req, userID)
		if err != nil {
			// Тут может быть как ошибка валидации урлов (bad request),
			// так и ошибка сохранения в стор (internal server error).
			// Можно в будущем добавить более детальную обработку ошибок, пока просто отдаём 400.
			utils.WriteJSONError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		utils.WriteJSON(w, http.StatusCreated, resp)
	}
}

func GetUserURLsHandler(svc *service.ShortenerService, up service.UserProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := RequireUserID(w, r, up)
		if !ok {
			return
		}

		urls, err := svc.GetUserURLs(r.Context(), userID)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "failed to get user URLs")
			return
		}

		if len(urls) == 0 {
			utils.WriteJSONError(w, http.StatusNoContent, http.StatusText(http.StatusNoContent))
			return
		}

		utils.WriteJSON(w, http.StatusOK, urls)
	}
}

func DeleteUserURLsHandler(svc *service.ShortenerService, up service.UserProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := RequireUserID(w, r, up)
		if !ok {
			return
		}

		var ids []string
		if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		if len(ids) == 0 {
			utils.WriteJSONError(w, http.StatusBadRequest, "no ids provided")
			return
		}

		svc.EnqueueDeletion(userID, ids)
		utils.WriteJSON(w, http.StatusAccepted, http.StatusText(http.StatusAccepted))
	}
}
