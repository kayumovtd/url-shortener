package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/kayumovtd/url-shortener/internal/model"
)

type InMemoryStore struct {
	mu      *sync.Mutex
	records []model.URLRecord
}

func (s *InMemoryStore) SaveURL(ctx context.Context, shortURL, originalURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.records {
		if rec.OriginalURL == originalURL {
			return NewErrStoreConflict(rec.ShortURL, rec.OriginalURL, nil)
		}
	}

	s.records = append(s.records, model.URLRecord{
		ID:          uuid.NewString(),
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})

	return nil
}

func (s *InMemoryStore) SaveURLs(ctx context.Context, urls map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for short, original := range urls {
		s.records = append(s.records, model.URLRecord{
			ID:          uuid.NewString(),
			UserID:      userID,
			ShortURL:    short,
			OriginalURL: original,
		})
	}

	return nil
}

func (s *InMemoryStore) GetURL(ctx context.Context, shortURL string) (model.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.records {
		if rec.ShortURL == shortURL {
			return rec, nil
		}
	}

	return model.URLRecord{}, errors.New("key not found")
}

func (s *InMemoryStore) GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	urls := []model.URLRecord{}
	for _, rec := range s.records {
		if rec.UserID == userID {
			urls = append(urls, rec)
		}
	}

	return urls, nil
}

func (s *InMemoryStore) MarkURLsDeleted(ctx context.Context, userID string, shortURLs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, rec := range s.records {
		for _, shortURL := range shortURLs {
			if rec.UserID == userID && rec.ShortURL == shortURL {
				s.records[i].IsDeleted = true
			}
		}
	}

	return nil
}

func (s *InMemoryStore) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Close() {}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		mu:      &sync.Mutex{},
		records: []model.URLRecord{},
	}
}
