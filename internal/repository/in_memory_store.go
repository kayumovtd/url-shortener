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

func (s *InMemoryStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = value
	return nil
}

func (s *InMemoryStore) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.store[key]
	if !ok {
		return "", errors.New("key not found")
	}

	return val, nil
}

func (s *InMemoryStore) Ping(ctx context.Context) error {
	return nil
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		mu:    &sync.Mutex{},
		store: make(map[string]string),
	}
}
