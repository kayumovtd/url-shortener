package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/service"
)

func TestShortenHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		decoded     bool
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
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				decoded:     true,
			},
		},
		{
			name: "invalid_json",
			body: `{"url":}`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				decoded:     false,
			},
		},
		{
			name: "empty_url",
			body: `{"url":""}`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				decoded:     false,
			},
		},
		{
			name: "invalid_url",
			body: `{"url":"fooBar"}`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				decoded:     false,
			},
		},
	}

	store := repository.NewInMemoryStore()
	svc := service.NewShortenerService(store, store, testBaseURL)
	handler := ShortenHandler(svc)

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

			if tt.want.contentType != "" {
				ct := res.Header.Get("Content-Type")
				if ct != tt.want.contentType {
					t.Errorf("content-type = %q, want %q", ct, tt.want.contentType)
				}
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
