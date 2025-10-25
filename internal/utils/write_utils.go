package utils

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kayumovtd/url-shortener/internal/model"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteJSONError(w http.ResponseWriter, status int, message string) {
	errResp := model.ErrorResponse{
		Code:    strconv.Itoa(status),
		Message: message,
	}
	WriteJSON(w, status, errResp)
}

func WritePlainText(w http.ResponseWriter, status int, v string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(v))
}
