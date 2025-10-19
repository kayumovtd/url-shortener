package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/service"
	"github.com/kayumovtd/url-shortener/internal/service/mocks"
)

const testBaseURL = "http://fooBar:8080"
const testUserID = "test_user_id"

func TestPostHandler(t *testing.T) {
	existingURL := "https://existing.com"
	existingShort := "abc123"

	type want struct {
		statusCode  int
		contentType string
		body        string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "valid_url",
			body: "https://example.com",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				body:        testBaseURL, // проверка, что ответ содержит какой-то урл
			},
		},
		{
			name: "empty_body",
			body: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
				body:        http.StatusText(http.StatusBadRequest),
			},
		},
		{
			name: "invalid_url",
			body: "fooBar",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
				body:        http.StatusText(http.StatusBadRequest),
			},
		},
		{
			name: "conflict_url",
			body: existingURL,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain",
				body:        testBaseURL, // проверка, что ответ содержит какой-то урл
			},
		},
	}

	store := repository.NewInMemoryStore()
	store.SaveURL(t.Context(), existingShort, existingURL, testUserID)

	svc := service.NewShortenerService(store, testBaseURL)
	up := mocks.NewMockUserProvider(testUserID, true)
	handler := PostHandler(svc, up)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			body := string(bodyBytes)

			if res.StatusCode != tt.want.statusCode {
				t.Fatalf("status code = %d, want %d", res.StatusCode, tt.want.statusCode)
			}

			gotCT := res.Header.Get("Content-Type")
			if gotCT != tt.want.contentType {
				t.Errorf("content type = %q, want %q", gotCT, tt.want.contentType)
			}

			if !strings.Contains(body, tt.want.body) {
				t.Errorf("body = %q, does not contain %q", body, tt.want.body)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	type want struct {
		statusCode int
		location   string
		body       string
	}

	tests := []struct {
		name string
		id   string
		want want
	}{
		{
			name: "existing_id",
			id:   "abc1",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example1.com",
			},
		},
		{
			name: "deleted_id",
			id:   "abc2",
			want: want{
				statusCode: http.StatusGone,
			},
		},
		{
			name: "non_existing_id",
			id:   "xyz999",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       http.StatusText(http.StatusBadRequest),
			},
		},
		{
			name: "empty_id",
			id:   "",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       http.StatusText(http.StatusBadRequest),
			},
		},
	}

	store := repository.NewMockStore()
	store.Data = []model.URLRecord{
		{ShortURL: "abc1", OriginalURL: "https://example1.com", UserID: testUserID, IsDeleted: false},
		{ShortURL: "abc2", OriginalURL: "https://example2.com", UserID: testUserID, IsDeleted: true},
	}

	svc := service.NewShortenerService(store, testBaseURL)
	handler := GetHandler(svc)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/" + tt.id
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			// в хендлере получаем параметр через chi, в тесте задать нужно через RouteContext
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			bodyBytes, _ := io.ReadAll(res.Body)
			body := string(bodyBytes)

			if res.StatusCode != tt.want.statusCode {
				t.Fatalf("status code = %d, want %d", res.StatusCode, tt.want.statusCode)
			}

			if tt.want.location != "" {
				gotLoc := res.Header.Get("Location")
				if gotLoc != tt.want.location {
					t.Errorf("location = %q, want %q", gotLoc, tt.want.location)
				}
			}

			if tt.want.body != "" && !strings.Contains(body, tt.want.body) {
				t.Errorf("body = %q, want %q", body, tt.want.body)
			}
		})
	}
}
