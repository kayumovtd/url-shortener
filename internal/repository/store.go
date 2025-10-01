package repository

import "context"

type Store interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Ping(ctx context.Context) error
}
