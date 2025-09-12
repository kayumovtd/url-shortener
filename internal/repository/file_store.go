package repository

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/kayumovtd/url-shortener/internal/model"
)

type FileStore struct {
	mu    *sync.Mutex
	store map[string]string
	path  string
}

func (s *FileStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = value
	return s.save()
}

func (s *FileStore) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.store[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (s *FileStore) save() error {
	records := make([]model.URLRecord, 0, len(s.store))
	i := 1
	for short, original := range s.store {
		records = append(records, model.URLRecord{
			// Пока просто используем индекс, т.к. ни на что не влияет,
			// потом можно заменить на UUID и хранить в сторе []URLRecord
			UUID:        strconv.Itoa(i),
			ShortURL:    short,
			OriginalURL: original,
		})
		i++
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{
		mu:    &sync.Mutex{},
		store: make(map[string]string),
		path:  path,
	}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var records []model.URLRecord
		if len(data) > 0 {
			if err := json.Unmarshal(data, &records); err != nil {
				return nil, err
			}
			for _, r := range records {
				fs.store[r.ShortURL] = r.OriginalURL
			}
		}
	}

	return fs, nil
}
