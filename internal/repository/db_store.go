package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/migrations"
)

type DBStore struct {
	pool *pgxpool.Pool
}

func (s *DBStore) SaveURL(ctx context.Context, shortURL, originalURL string, userID string) error {
	query := `
		INSERT INTO urls (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
		RETURNING short_url;
	`

	var result string
	err := s.pool.QueryRow(ctx, query, shortURL, originalURL, userID).Scan(&result)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return NewErrStoreConflict(shortURL, originalURL, err)
		}
		return err
	}
	return nil
}

func (s *DBStore) SaveURLs(ctx context.Context, urls map[string]string, userID string) error {
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
			`INSERT INTO urls (short_url, original_url, user_id)
			 VALUES ($1, $2, $3) 
		 	 ON CONFLICT (short_url) DO UPDATE SET original_url = EXCLUDED.original_url`,
			short, orig, userID,
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

func (s *DBStore) GetURL(ctx context.Context, shortURL string) (model.URLRecord, error) {
	var result model.URLRecord
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, short_url, original_url, is_deleted FROM urls WHERE short_url = $1`,
		shortURL,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.ShortURL,
		&result.OriginalURL,
		&result.IsDeleted,
	)

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return model.URLRecord{}, fmt.Errorf("request canceled or timed out: %w", err)
	}
	if err != nil {
		return model.URLRecord{}, fmt.Errorf("failed to get original url: %w", err)
	}
	return result, nil
}

func (s *DBStore) GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, short_url, original_url FROM urls WHERE user_id = $1`,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query user urls: %w", err)
	}
	defer rows.Close()

	urls := []model.URLRecord{}
	for rows.Next() {
		var record model.URLRecord
		if err := rows.Scan(&record.ID, &record.ShortURL, &record.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		urls = append(urls, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return urls, nil
}

func (s *DBStore) MarkURLsDeleted(ctx context.Context, userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	query := `
        UPDATE urls
        SET is_deleted = TRUE
        WHERE user_id = $1 AND short_url = ANY($2)
    `
	_, err := s.pool.Exec(ctx, query, userID, shortURLs)
	if err != nil {
		return fmt.Errorf("failed to mark urls deleted: %w", err)
	}

	return nil
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
