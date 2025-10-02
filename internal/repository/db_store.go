package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kayumovtd/url-shortener/migrations"
)

type DBStore struct {
	pool *pgxpool.Pool
}

func (s *DBStore) SaveURL(ctx context.Context, shortURL, originalURL string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO urls (short_url, original_url) 
		 VALUES ($1, $2)
		 ON CONFLICT (short_url) DO UPDATE SET original_url = EXCLUDED.original_url`,
		shortURL, originalURL,
	)
	if err != nil {
		return fmt.Errorf("failed to set url: %w", err)
	}
	return nil
}

func (s *DBStore) SaveURLs(ctx context.Context, urls map[string]string) error {
	if len(urls) == 0 {
		return nil
	}

	const batchSize = 500

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	count := 0

	for short, orig := range urls {
		batch.Queue(
			`INSERT INTO urls (short_url, original_url)
			 VALUES ($1, $2) 
		 	 ON CONFLICT (short_url) DO UPDATE SET original_url = EXCLUDED.original_url`,
			short, orig,
		)
		count++

		if count >= batchSize {
			br := tx.SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				return fmt.Errorf("batch execution failed: %w", err)
			}

			batch = &pgx.Batch{}
			count = 0
		}
	}

	// финальный батч (если что-то осталось < batchSize)
	if count > 0 {
		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return fmt.Errorf("final batch execution failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (s *DBStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	var original string
	err := s.pool.QueryRow(ctx,
		`SELECT original_url FROM urls WHERE short_url = $1`,
		shortURL,
	).Scan(&original)

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("request canceled or timed out: %w", err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get original url: %w", err)
	}
	return original, nil
}

func (s *DBStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *DBStore) Close() {
	s.pool.Close()
}

func NewDBStore(dsn string) (*DBStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if err := migrations.ApplyMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &DBStore{pool: pool}, nil
}
