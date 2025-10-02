package repository

import (
	"context"
	"errors"
	"maps"
)

type MockErrorType int

const (
	NoError MockErrorType = iota
	SomeError
	ConflictError
)

// TODO: Заюзать gomock
type MockStore struct {
	Data      map[string]string
	ErrorType MockErrorType
}

func NewMockStore() *MockStore {
	return &MockStore{Data: make(map[string]string)}
}

func (f *MockStore) SaveURL(ctx context.Context, shortURL, originalURL string) error {
	switch f.ErrorType {
	case NoError:
		// continue
	case SomeError:
		return errors.New("some error")
	case ConflictError:
		return NewErrStoreConflict(shortURL, originalURL, errors.New("conflict"))
	}

	f.Data[shortURL] = originalURL
	return nil
}

func (f *MockStore) SaveURLs(ctx context.Context, urls map[string]string) error {
	switch f.ErrorType {
	case NoError:
		// continue
	case SomeError:
		return errors.New("some error")
	case ConflictError:
		return NewErrStoreConflict("", "", errors.New("conflict"))
	}
	maps.Copy(f.Data, urls)
	return nil
}

func (f *MockStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	val, ok := f.Data[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	return val, nil
}

func (f *MockStore) Ping(ctx context.Context) error {
	return nil
}

func (f *MockStore) Close() {}
