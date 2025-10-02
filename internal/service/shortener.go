package service

import (
	"context"
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

func (s *ShortenerService) Shorten(ctx context.Context, originalURL string) (string, error) {
	trimmedURL := strings.TrimSpace(originalURL)

	u, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", originalURL, err)
	}
	normalized := u.String()

	shortID := utils.GenerateID(normalized)

	if err := s.store.SaveURL(ctx, shortID, normalized); err != nil {
		return "", fmt.Errorf("failed to save url %q with id %q: %w", normalized, shortID, err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, shortID), nil
}

func (s *ShortenerService) Unshorten(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("empty id")
	}

	orig, err := s.store.GetURL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("not found: %w", err)
	}

	return orig, nil
}

func (s *ShortenerService) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}
