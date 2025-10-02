package service

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/kayumovtd/url-shortener/internal/model"
)

// TODO: Заюзать gomock
type mockStore struct {
	data map[string]string
	fail bool
}

func newFakeStore() *mockStore {
	return &mockStore{data: make(map[string]string)}
}

func (f *mockStore) SaveURL(ctx context.Context, shortURL, originalURL string) error {
	if f.fail {
		return errors.New("store error")
	}
	f.data[shortURL] = originalURL
	return nil
}

func (f *mockStore) SaveURLs(ctx context.Context, urls map[string]string) error {
	if f.fail {
		return errors.New("store error")
	}
	maps.Copy(f.data, urls)
	return nil
}

func (f *mockStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	val, ok := f.data[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	return val, nil
}

func (f *mockStore) Ping(ctx context.Context) error {
	return nil
}

func (f *mockStore) Close() {}

const testBaseURL = "http://fooBar:8080"

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

	store := newFakeStore()
	svc := NewShortenerService(store, testBaseURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Shorten(t.Context(), tt.input)
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

	store := newFakeStore()
	svc := NewShortenerService(store, testBaseURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ShortenBatch(context.Background(), tt.input)

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

	store := newFakeStore()
	svc := NewShortenerService(store, testBaseURL)

	_ = store.SaveURL(t.Context(), "fooBar", "https://example.com")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Unshorten(t.Context(), tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil (result=%q)", got)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestShorten_StoreError(t *testing.T) {
	store := newFakeStore()
	store.fail = true

	svc := NewShortenerService(store, testBaseURL)

	_, err := svc.Shorten(t.Context(), "https://example.com")
	if err == nil {
		t.Errorf("expected store error, got nil")
	}
}
