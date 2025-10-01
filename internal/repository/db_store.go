package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStore struct {
	pool *pgxpool.Pool
}

func NewDBStore(dsn string) (*DBStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %w", err)
	}

	// if err := pool.Ping(ctx); err != nil {
	// 	return nil, fmt.Errorf("failed to ping postgres: %w", err)
	// }

	return &DBStore{pool: pool}, nil
}

func (s *DBStore) Set(key, value string) error {
	return fmt.Errorf("not implemented yet")
}

func (s *DBStore) Get(key string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (s *DBStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
