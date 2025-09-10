package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"data":"` + string(data) + `"}`))
}

func TestGzipMiddleware(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(mockHandler))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{"msg":"hello"}`
	successBody := `{"data":"` + requestBody + `"}`

	t.Run("sends_gzip", func(t *testing.T) {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write([]byte(requestBody)); err != nil {
			t.Fatalf("gzip write failed: %v", err)
		}
		if err := zw.Close(); err != nil {
			t.Fatalf("gzip close failed: %v", err)
		}

		req := httptest.NewRequest("POST", srv.URL, &buf)
		req.RequestURI = ""
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusOK)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}

		if string(data) != successBody {
			t.Fatalf("unexpected body:\ngot  %s\nwant %s", string(data), successBody)
		}
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		req := httptest.NewRequest("POST", srv.URL, strings.NewReader(requestBody))
		req.RequestURI = ""
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusOK)
		}

		zr, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatalf("gzip reader failed: %v", err)
		}
		defer zr.Close()

		data, err := io.ReadAll(zr)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}

		if string(data) != successBody {
			t.Fatalf("unexpected body:\ngot  %s\nwant %s", string(data), successBody)
		}
	})
}
