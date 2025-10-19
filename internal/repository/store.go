package repository

import (
	"context"

	"github.com/kayumovtd/url-shortener/internal/model"
)

type Store interface {
	SaveURL(ctx context.Context, shortURL, originalURL string, userID string) error
	SaveURLs(ctx context.Context, urls map[string]string, userID string) error
	GetURL(ctx context.Context, shortURL string) (model.URLRecord, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error)
	Ping(ctx context.Context) error
	Close()
}
