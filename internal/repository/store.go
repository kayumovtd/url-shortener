package repository

import (
	"context"
)

type Store interface {
	SaveURL(ctx context.Context, shortURL, originalURL string) error
	SaveURLs(ctx context.Context, urls map[string]string) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	Close()
}
