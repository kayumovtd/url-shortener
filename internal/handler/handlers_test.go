package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/service"
)

const testBaseURL = "http://fooBar:8080"

func TestPostHandler(t *testing.T) {
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
				contentType: "text/plain; charset=utf-8",
				body:        http.StatusText(http.StatusBadRequest),
			},
		},
		{
			name: "invalid_url",
			body: "fooBar",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        http.StatusText(http.StatusBadRequest),
			},
		},
	}

	store := repository.NewInMemoryStore()
	svc := service.NewShortenerService(store, testBaseURL)
	handler := PostHandler(svc)

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
		name         string
		id           string
		prepareStore func(ctx context.Context, store repository.Store)
		want         want
	}{
		{
			name: "existing_id",
			id:   "abc123",
			prepareStore: func(ctx context.Context, store repository.Store) {
				store.SaveURL(ctx, "abc123", "https://example.com")
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com",
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

	for _, tt := range tests {
		store := repository.NewInMemoryStore()
		if tt.prepareStore != nil {
			tt.prepareStore(t.Context(), store)
		}
		svc := service.NewShortenerService(store, testBaseURL)
		handler := GetHandler(svc)

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
