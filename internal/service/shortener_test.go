package service

import (
	"context"
	"errors"
	"testing"
)

type mockStore struct {
	data map[string]string
	fail bool
}

func newFakeStore() *mockStore {
	return &mockStore{data: make(map[string]string)}
}

func (f *mockStore) Set(id, url string) error {
	if f.fail {
		return errors.New("store error")
	}
	f.data[id] = url
	return nil
}

func (f *mockStore) Get(id string) (string, error) {
	val, ok := f.data[id]
	if !ok {
		return "", errors.New("not found")
	}
	return val, nil
}

func (f *mockStore) Ping(ctx context.Context) error {
	return nil
}

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
	svc := NewShortenerService(store, store, testBaseURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Shorten(tt.input)
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
	svc := NewShortenerService(store, store, testBaseURL)

	_ = store.Set("fooBar", "https://example.com")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Unshorten(tt.input)

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

	svc := NewShortenerService(store, store, testBaseURL)

	_, err := svc.Shorten("https://example.com")
	if err == nil {
		t.Errorf("expected store error, got nil")
	}
}
