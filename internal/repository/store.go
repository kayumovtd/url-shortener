package repository

import "sync"

type Store interface {
	Set(key, value string) error
	Get(key string) (string, bool)
}

type InMemoryStore struct {
	mu    sync.RWMutex
	store map[string]string
}

func (s *InMemoryStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
	return nil
}

func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[key]
	return val, ok
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store: make(map[string]string),
	}
}
