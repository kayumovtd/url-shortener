package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
				body:        "http://", // проверка, что ответ содержит какой-то урл
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

	for _, tt := range tests {
		store := map[string]string{}
		handler := PostHandler(store)

		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			body := string(bodyBytes)

			if res.StatusCode != tt.want.statusCode {
				t.Errorf("status code = %d, want %d", res.StatusCode, tt.want.statusCode)
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
		name  string
		id    string
		store map[string]string
		want  want
	}{
		{
			name:  "existing_id",
			id:    "abc123",
			store: map[string]string{"abc123": "https://example.com"},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com",
				body:       "",
			},
		},
		{
			name:  "non-existing_id",
			id:    "xyz999",
			store: map[string]string{},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
				body:       http.StatusText(http.StatusBadRequest),
			},
		},
		{
			name:  "empty_id",
			id:    "",
			store: map[string]string{},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
				body:       http.StatusText(http.StatusBadRequest),
			},
		},
	}

	for _, tt := range tests {
		handler := GetHandler(tt.store)

		t.Run(tt.name, func(t *testing.T) {
			url := "/" + tt.id
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			body := string(bodyBytes)

			if res.StatusCode != tt.want.statusCode {
				t.Errorf("status code = %d, want %d", res.StatusCode, tt.want.statusCode)
			}

			if tt.want.location != "" {
				gotLoc := res.Header.Get("Location")
				if gotLoc != tt.want.location {
					t.Errorf("location = %q, want %q", gotLoc, tt.want.location)
				}
			}

			if tt.want.body != "" && !strings.Contains(body, tt.want.body) {
				t.Errorf("body = %q, does not contain %q", body, tt.want.body)
			}
		})
	}
}
