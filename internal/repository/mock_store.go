package repository

import (
	"context"
	"errors"

	"github.com/kayumovtd/url-shortener/internal/model"
)

type MockErrorType int

const (
	NoError MockErrorType = iota
	SomeError
	ConflictError
)

// TODO: Заюзать gomock
type MockStore struct {
	Data      []model.URLRecord
	ErrorType MockErrorType
}

func NewMockStore() *MockStore {
	return &MockStore{}
}

func (f *MockStore) SaveURL(ctx context.Context, shortURL, originalURL, userID string) error {
	switch f.ErrorType {
	case NoError:
		// continue
	case SomeError:
		return errors.New("some error")
	case ConflictError:
		return NewErrStoreConflict(shortURL, originalURL, errors.New("conflict"))
	}
	return nil
}

func (f *MockStore) SaveURLs(ctx context.Context, urls map[string]string, userID string) error {
	switch f.ErrorType {
	case NoError:
		// continue
	case SomeError:
		return errors.New("some error")
	case ConflictError:
		return NewErrStoreConflict("", "", errors.New("conflict"))
	}
	return nil
}

func (f *MockStore) GetURL(ctx context.Context, shortURL string) (model.URLRecord, error) {
	for _, rec := range f.Data {
		if rec.ShortURL == shortURL {
			return rec, nil
		}
	}

	return model.URLRecord{}, errors.New("key not found")
}

func (f *MockStore) GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	urls := []model.URLRecord{}
	for _, rec := range f.Data {
		if rec.UserID == userID {
			urls = append(urls, rec)
		}
	}

	return urls, nil
}

func (f *MockStore) MarkURLsDeleted(ctx context.Context, userID string, shortURLs []string) error {
	for i, rec := range f.Data {
		for _, shortURL := range shortURLs {
			if rec.UserID == userID && rec.ShortURL == shortURL {
				f.Data[i].IsDeleted = true
			}
		}
	}

	return nil
}

func (f *MockStore) Ping(ctx context.Context) error {
	return nil
}

func (f *MockStore) Close() {}
