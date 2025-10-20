package handler

import (
	"net/http"

	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

func RequireUserID(w http.ResponseWriter, r *http.Request, up service.UserProvider) (string, bool) {
	userID, ok := up.GetUserID(r.Context())
	if !ok || userID == "" {
		utils.WritePlainText(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return "", false
	}
	return userID, true
}
