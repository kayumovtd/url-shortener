package repository

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("key not found")
)

type Store interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

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
		return "", ErrNotFound
	}

	return val, nil
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		mu:    &sync.Mutex{},
		store: make(map[string]string),
	}
}
