package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/service/mocks"
)

func TestShortenHandler(t *testing.T) {
	existingURL := "https://existing.com"
	existingShort := "abc123"

	type want struct {
		statusCode int
		decoded    bool
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "valid_request",
			body: `{"url":"https://example.com"}`,
			want: want{
				statusCode: http.StatusCreated,
				decoded:    true,
			},
		},
		{
			name: "invalid_json",
			body: `{"url":}`,
			want: want{
				statusCode: http.StatusBadRequest,
				decoded:    false,
			},
		},
		{
			name: "empty_url",
			body: `{"url":""}`,
			want: want{
				statusCode: http.StatusBadRequest,
				decoded:    false,
			},
		},
		{
			name: "invalid_url",
			body: `{"url":"fooBar"}`,
			want: want{
				statusCode: http.StatusBadRequest,
				decoded:    false,
			},
		},
		{
			name: "conflict_url",
			body: fmt.Sprintf(`{"url":"%s"}`, existingURL),
			want: want{
				statusCode: http.StatusConflict,
				decoded:    true,
			},
		},
	}

	store := repository.NewInMemoryStore()
	store.SaveURL(t.Context(), existingShort, existingURL, testUserID)

	svc := service.NewShortenerService(store, testBaseURL)
	up := mocks.NewMockUserProvider(testUserID, true)
	handler := ShortenHandler(svc, up)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.statusCode {
				t.Errorf("status = %d, want %d", res.StatusCode, tt.want.statusCode)
			}

			if tt.want.decoded {
				var resp model.ShortenResponse
				if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
					t.Errorf("response is not valid JSON: %v", err)
				}
				if resp.Result == "" {
					t.Errorf("expected non-empty result")
				}
			}
		})
	}
}

func TestGetUserURLsHandler(t *testing.T) {
	type want struct {
		statusCode int
		decoded    bool
	}

	tests := []struct {
		name   string
		userID string
		want   want
	}{
		{
			name:   "has_urls",
			userID: testUserID,
			want: want{
				statusCode: http.StatusOK,
				decoded:    true,
			},
		},
		{
			name:   "unauthorized",
			userID: "",
			want: want{
				statusCode: http.StatusUnauthorized,
				decoded:    false,
			},
		},
		{
			name:   "no_content",
			userID: "some_other_user",
			want: want{
				statusCode: http.StatusNoContent,
				decoded:    false,
			},
		},
	}

	store := repository.NewMockStore()
	store.Data = []model.URLRecord{
		{ShortURL: "fooBar1", OriginalURL: "https://example.com", UserID: testUserID},
		{ShortURL: "fooBar2", OriginalURL: "https://example.com", UserID: testUserID},
	}
	svc := service.NewShortenerService(store, testBaseURL)

	for _, tt := range tests {
		up := mocks.NewMockUserProvider(tt.userID, true)
		handler := GetUserURLsHandler(svc, up)

		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.statusCode {
				t.Errorf("status = %d, want %d", res.StatusCode, tt.want.statusCode)
			}

			if tt.want.decoded {
				var resp []model.UserURLsResponseItem
				if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
					t.Errorf("response is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestDeleteUserURLsHandler(t *testing.T) {
	store := repository.NewMockStore()
	svc := service.NewShortenerService(store, testBaseURL)
	up := mocks.NewMockUserProvider(testUserID, true)
	handler := DeleteUserURLsHandler(svc, up)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBufferString(`["id1","id2","id3"]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusAccepted)
	}
}
