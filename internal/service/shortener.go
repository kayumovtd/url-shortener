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
	store        repository.Store
	baseURL      string
	batchDeleter *BatchDeleter
}

func NewShortenerService(store repository.Store, baseURL string, batchDeleter *BatchDeleter) *ShortenerService {
	return &ShortenerService{
		store:        store,
		baseURL:      baseURL,
		batchDeleter: batchDeleter,
	}
}

func (s *ShortenerService) Shorten(ctx context.Context, originalURL string, userID string) (string, error) {
	url, err := s.normalizeURL(originalURL)
	if err != nil {
		return "", err
	}

	shortID := utils.GenerateID(url)

	if err := s.store.SaveURL(ctx, shortID, url, userID); err != nil {
		var conflict *repository.ErrStoreConflict
		if errors.As(err, &conflict) {
			return "", NewErrShortenerConflict(s.makeResultURL(conflict.ShortURL), err)
		}
		return "", fmt.Errorf("failed to save url %q with id %q: %w", url, shortID, err)
	}

	return s.makeResultURL(shortID), nil
}

func (s *ShortenerService) ShortenBatch(
	ctx context.Context,
	items []model.ShortenBatchRequestItem,
	userID string,
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
		// TODO: Можно сделать свой тип ошибки и выдавать юзеру массив битых урлов
		return nil, errors.Join(errs...)
	}

	if err := s.store.SaveURLs(ctx, pairs, userID); err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	return responses, nil
}

func (s *ShortenerService) Unshorten(ctx context.Context, id string) (model.URLRecord, error) {
	if id == "" {
		return model.URLRecord{}, fmt.Errorf("empty id")
	}

	rec, err := s.store.GetURL(ctx, id)
	if err != nil {
		return model.URLRecord{}, fmt.Errorf("not found: %w", err)
	}

	return rec, nil
}

func (s *ShortenerService) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}

func (s *ShortenerService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURLsResponseItem, error) {
	urls, err := s.store.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	response := make([]model.UserURLsResponseItem, 0, len(urls))
	for _, rec := range urls {
		response = append(response, model.UserURLsResponseItem{
			ShortURL:    s.makeResultURL(rec.ShortURL),
			OriginalURL: rec.OriginalURL,
		})
	}

	return response, nil
}

func (s *ShortenerService) normalizeURL(originalURL string) (string, error) {
	trimmedURL := strings.TrimSpace(originalURL)

	u, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", originalURL, err)
	}

	return u.String(), nil
}

func (s *ShortenerService) makeResultURL(shortID string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, shortID)
}

func (s *ShortenerService) EnqueueDeletion(userID string, shortIDs []string) {
	s.batchDeleter.Enqueue(userID, shortIDs)
}
