package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/kayumovtd/url-shortener/internal/model"
)

type FileStore struct {
	mu      *sync.Mutex
	records []model.URLRecord
	path    string
}

func (s *FileStore) SaveURL(ctx context.Context, shortURL, originalURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.records {
		if rec.OriginalURL == originalURL {
			return NewErrStoreConflict(rec.ShortURL, rec.OriginalURL, nil)
		}
	}

	record := model.URLRecord{
		ID:          uuid.NewString(),
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	s.records = append(s.records, record)
	return s.save()
}

func (s *FileStore) SaveURLs(ctx context.Context, urls map[string]string, userID string) error {
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

	return s.save()
}

func (s *FileStore) GetURL(ctx context.Context, shortURL string) (model.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.records {
		if rec.ShortURL == shortURL {
			return rec, nil
		}
	}

	return model.URLRecord{}, errors.New("key not found")
}

func (s *FileStore) GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
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

func (s *FileStore) MarkURLsDeleted(ctx context.Context, userID string, shortURLs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, rec := range s.records {
		for _, shortURL := range shortURLs {
			if rec.UserID == userID && rec.ShortURL == shortURL {
				s.records[i].IsDeleted = true
			}
		}
	}

	return s.save()
}

func (s *FileStore) save() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *FileStore) Ping(ctx context.Context) error {
	return nil
}

func (s *FileStore) Close() {}

func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{
		mu:      &sync.Mutex{},
		records: []model.URLRecord{},
		path:    path,
	}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &fs.records); err != nil {
				return nil, err
			}
		}
	}

	return fs, nil
}
