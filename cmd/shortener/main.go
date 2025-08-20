package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const shortIDLength = 8 // длина укороченного идентификатора

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func generateID(url string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	if shortIDLength > len(encoded) {
		return encoded
	}
	return encoded[:shortIDLength]
}

func handlePost(w http.ResponseWriter, r *http.Request) {
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

	shortID := generateID(originalURL)
	shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortID)

	mu.Lock()
	store[shortID] = originalURL
	mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func handleGet(w http.ResponseWriter, r *http.Request) {
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

func main() {
	http.HandleFunc("/", handlePost)
	http.HandleFunc("/{id}", handleGet)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
