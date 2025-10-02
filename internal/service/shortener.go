package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/kayumovtd/url-shortener/internal/model"
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
	url, err := s.normalizeURL(originalURL)
	if err != nil {
		return "", err
	}

	shortID := utils.GenerateID(url)

	if err := s.store.SaveURL(ctx, shortID, url); err != nil {
		return "", fmt.Errorf("failed to save url %q with id %q: %w", url, shortID, err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, shortID), nil
}

func (s *ShortenerService) ShortenBatch(
	ctx context.Context,
	items []model.ShortenBatchRequestItem,
) ([]model.ShortenBatchResponseItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	pairs := make(map[string]string, len(items))
	responses := make([]model.ShortenBatchResponseItem, 0, len(items))

	var errs []error
	hasErrors := false

	for _, it := range items {
		normalized, err := s.normalizeURL(it.OriginalURL)
		// Условимся, что если хоть в одном из элементов батча ошибка, то весь батч не сохраняем
		if err != nil {
			errs = append(errs, fmt.Errorf("correlation_id %q: %w", it.CorrelationID, err))
			hasErrors = true
			continue
		}

		// Если ошибки уже были, shortID даже не генерируем
		if hasErrors {
			continue
		}

		shortID := utils.GenerateID(normalized)
		pairs[shortID] = normalized

		responses = append(responses, model.ShortenBatchResponseItem{
			CorrelationID: it.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, shortID),
		})
	}

	if hasErrors {
		// Отдаём все ошибки через Join
		return nil, errors.Join(errs...)
	}

	if err := s.store.SaveURLs(ctx, pairs); err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	return responses, nil
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

func (s *ShortenerService) normalizeURL(originalURL string) (string, error) {
	trimmedURL := strings.TrimSpace(originalURL)

	u, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", originalURL, err)
	}

	return u.String(), nil
}
