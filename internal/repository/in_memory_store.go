package repository

import (
	"context"
	"errors"
	"sync"
)

type InMemoryStore struct {
	mu    *sync.Mutex
	store map[string]string
}

func (s *InMemoryStore) SaveURL(ctx context.Context, shortURL, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[shortURL] = originalURL
	return nil
}

func (s *InMemoryStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.store[shortURL]
	if !ok {
		return "", errors.New("key not found")
	}

	return val, nil
}

func (s *InMemoryStore) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Close() {}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		mu:    &sync.Mutex{},
		store: make(map[string]string),
	}
}
