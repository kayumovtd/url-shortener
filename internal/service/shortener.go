package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/utils"
)

type ShortenerService struct {
	store   repository.Store
	baseURL string
}

func NewShortenerService(store repository.Store, baseURL string) *ShortenerService {
	return &ShortenerService{store: store, baseURL: baseURL}
}

func (s *ShortenerService) Shorten(originalURL string) (string, error) {
	trimmedURL := strings.TrimSpace(originalURL)

	u, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", originalURL, err)
	}
	normalized := u.String()

	shortID := utils.GenerateID(normalized)

	if err := s.store.Set(shortID, normalized); err != nil {
		return "", fmt.Errorf("failed to save url %q with id %q: %w", normalized, shortID, err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, shortID), nil
}

func (s *ShortenerService) Unshorten(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("empty id")
	}

	orig, err := s.store.Get(id)
	if err != nil {
		return "", fmt.Errorf("not found: %w", err)
	}

	return orig, nil
}
