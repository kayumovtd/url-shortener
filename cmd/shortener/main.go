package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/kayumovtd/url-shortener/internal/utils"
)

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func PostHandler(store map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		originalURL := strings.TrimSpace(string(body))
		u, err := url.ParseRequestURI(originalURL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		originalURL = u.String()
		shortID := utils.GenerateID(originalURL)
		shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortID)

		mu.Lock()
		store[shortID] = originalURL
		mu.Unlock()

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func GetHandler(store map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		mu.RLock()
		origURL, ok := store[id]
		mu.RUnlock()

		if !ok {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
	}
}

func main() {
	http.HandleFunc("/", PostHandler(store))
	http.HandleFunc("/{id}", GetHandler(store))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
