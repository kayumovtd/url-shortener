package service

import (
	"testing"
	"time"

	"github.com/kayumovtd/url-shortener/internal/model"
	"github.com/kayumovtd/url-shortener/internal/repository"
)

func TestBatchDeleter_FlushOnBatchSize(t *testing.T) {
	mockStore := repository.NewMockStore()
	mockStore.Data = []model.URLRecord{
		{ShortURL: "a", UserID: "user1", IsDeleted: false},
		{ShortURL: "b", UserID: "user1", IsDeleted: false},
		{ShortURL: "c", UserID: "user1", IsDeleted: false},
	}

	deleter := NewBatchDeleter(mockStore)
	defer deleter.Close()

	deleter.SetBatchSize(3)
	deleter.SetFlushInterval(10 * time.Second) // Чтобы не сработал таймер

	// Отправляем 3 задания
	deleter.Enqueue("user1", []string{"a"})
	deleter.Enqueue("user1", []string{"b"})
	deleter.Enqueue("user1", []string{"c"})

	time.Sleep(100 * time.Millisecond) // Даём время на обработку

	var deleted int
	for _, rec := range mockStore.Data {
		if rec.IsDeleted {
			deleted++
		}
	}

	if deleted != 3 {
		t.Fatalf("expected 3 deleted ids, got %d", deleted)
	}
}

func TestBatchDeleter_FlushOnTimer(t *testing.T) {
	mockStore := repository.NewMockStore()
	mockStore.Data = []model.URLRecord{
		{ShortURL: "a", UserID: "user1", IsDeleted: false},
		{ShortURL: "b", UserID: "user1", IsDeleted: false},
		{ShortURL: "c", UserID: "user1", IsDeleted: false},
	}

	deleter := NewBatchDeleter(mockStore)
	defer deleter.Close()

	deleter.SetBatchSize(100) // Чтобы не сработал батч
	deleter.SetFlushInterval(100 * time.Millisecond)

	// Отправляем 1 задание
	deleter.Enqueue("user1", []string{"a"})
	time.Sleep(200 * time.Millisecond) // Ждём flush по таймеру

	var deleted int
	for _, rec := range mockStore.Data {
		if rec.IsDeleted {
			deleted++
		}
	}

	if deleted != 1 {
		t.Fatalf("expected 1 deleted ids, got %d", deleted)
	}
}
