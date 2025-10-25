package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kayumovtd/url-shortener/internal/logger"
	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/repository"
)

const testBaseURL = "http://fooBar:8080"
const testUserID = "test_user_id"

func TestShorten(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid_url", "https://example.com", false},
		{"invalid_url", "foobar", true},
		{"with_spaces", "   https://example.com   ", false},
		{"empty_string", "", true},
	}

	store := repository.NewMockStore()
	bd := NewBatchDeleter(store, logger.NewNoOp())
	defer bd.Close()
	svc := NewShortenerService(store, testBaseURL, bd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Shorten(t.Context(), tt.input, testUserID)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil (result=%q)", got)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got == "" {
				t.Errorf("expected non-empty short url")
			}
		})
	}
}

func TestShortenBatch(t *testing.T) {
	tests := []struct {
		name      string
		input     []model.ShortenBatchRequestItem
		shouldErr bool
	}{
		{
			name: "valid_batch",
			input: []model.ShortenBatchRequestItem{
				{CorrelationID: "1", OriginalURL: "https://example1.com"},
				{CorrelationID: "2", OriginalURL: "https://example2.com"},
			},
			shouldErr: false,
		},
		{
			name: "invalid_batch",
			input: []model.ShortenBatchRequestItem{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
				{CorrelationID: "2", OriginalURL: "foobar"},
			},
			shouldErr: true,
		},
		{
			name:      "empty_batch",
			input:     []model.ShortenBatchRequestItem{},
			shouldErr: true,
		},
	}

	store := repository.NewMockStore()
	bd := NewBatchDeleter(store, logger.NewNoOp())
	defer bd.Close()
	svc := NewShortenerService(store, testBaseURL, bd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ShortenBatch(context.Background(), tt.input, testUserID)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil (result=%v)", got)
				}
				if got != nil {
					t.Errorf("expected nil result, got: %v", got)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(got) != len(tt.input) {
				t.Errorf("expected %d results, got %d", len(tt.input), len(got))
			}
			for _, resp := range got {
				if resp.ShortURL == "" {
					t.Errorf("expected non-empty short url for correlation_id=%s", resp.CorrelationID)
				}
			}
		})
	}
}

func TestUnshorten(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		shouldErr bool
	}{
		{"valid_id", "fooBar", "https://example.com", false},
		{"unknown_id", "xxx", "", true},
		{"empty_id", "", "", true},
	}

	store := repository.NewMockStore()
	store.Data = []model.URLRecord{
		{ShortURL: "fooBar", OriginalURL: "https://example.com"},
	}
	bd := NewBatchDeleter(store, logger.NewNoOp())
	defer bd.Close()
	svc := NewShortenerService(store, testBaseURL, bd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Unshorten(t.Context(), tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil (result=%q)", got.OriginalURL)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got.OriginalURL != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got.OriginalURL)
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		wantURLs bool
	}{
		{"has_urls", testUserID, true},
		{"has_no_urls", "some_other_user", false},
	}

	store := repository.NewMockStore()
	store.Data = []model.URLRecord{
		{ShortURL: "fooBar1", OriginalURL: "https://example.com", UserID: testUserID},
		{ShortURL: "fooBar2", OriginalURL: "https://example.com", UserID: testUserID},
	}
	bd := NewBatchDeleter(store, logger.NewNoOp())
	defer bd.Close()
	svc := NewShortenerService(store, testBaseURL, bd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetUserURLs(t.Context(), tt.userID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			hasURLs := len(got) > 0
			if hasURLs != tt.wantURLs {
				t.Errorf("expected %v, got %v", tt.wantURLs, hasURLs)
			}
		})
	}
}

func TestShorten_StoreError(t *testing.T) {
	store := repository.NewMockStore()
	store.ErrorType = repository.SomeError

	bd := NewBatchDeleter(store, logger.NewNoOp())
	defer bd.Close()
	svc := NewShortenerService(store, testBaseURL, bd)

	_, err := svc.Shorten(t.Context(), "https://example.com", testUserID)
	if err == nil {
		t.Errorf("expected store error, got nil")
	}
}

func TestShorten_ConflictStoreError(t *testing.T) {
	store := repository.NewMockStore()
	store.ErrorType = repository.ConflictError

	bd := NewBatchDeleter(store, logger.NewNoOp())
	defer bd.Close()
	svc := NewShortenerService(store, testBaseURL, bd)

	_, err := svc.Shorten(t.Context(), "https://example.com", testUserID)
	if err == nil {
		t.Errorf("expected store error, got nil")
	}
	var conflictErr *ErrShortenerConflict
	if !errors.As(err, &conflictErr) {
		t.Errorf("expected conflict store error, got: %v", err)
	}
}
