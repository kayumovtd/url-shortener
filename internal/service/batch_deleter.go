package service

import (
	"context"
	"time"

	"github.com/kayumovtd/url-shortener/internal/repository"
)

type DeleteTask struct {
	UserID   string
	ShortIDs []string
}

type BatchDeleter struct {
	store         repository.Store
	inputCh       chan DeleteTask
	doneCh        chan struct{}
	batchSize     int
	flushInterval time.Duration
}

func NewBatchDeleter(store repository.Store) *BatchDeleter {
	h := &BatchDeleter{
		store:         store,
		inputCh:       make(chan DeleteTask, 1000),
		doneCh:        make(chan struct{}),
		batchSize:     100,
		flushInterval: 2 * time.Second,
	}

	go h.runAggregator()

	return h
}

func (h *BatchDeleter) runAggregator() {
	batch := make([]DeleteTask, 0, h.batchSize)
	ticker := time.NewTicker(h.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case task := <-h.inputCh:
			batch = append(batch, task)
			if len(batch) >= h.batchSize {
				h.flush(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				h.flush(batch)
				batch = batch[:0]
			}

		case <-h.doneCh:
			if len(batch) > 0 {
				h.flush(batch)
			}
			return
		}
	}
}

func (h *BatchDeleter) flush(batch []DeleteTask) {
	userURLs := make(map[string][]string)
	for _, t := range batch {
		userURLs[t.UserID] = append(userURLs[t.UserID], t.ShortIDs...)
	}

	for userID, ids := range userURLs {
		_ = h.store.MarkURLsDeleted(context.Background(), userID, ids)
	}
}

func (h *BatchDeleter) Enqueue(userID string, shortIDs []string) {
	h.inputCh <- DeleteTask{UserID: userID, ShortIDs: shortIDs}
}

func (h *BatchDeleter) Close() {
	close(h.inputCh)
}

func (h *BatchDeleter) SetBatchSize(size int) {
	h.batchSize = size
}

func (h *BatchDeleter) SetFlushInterval(d time.Duration) {
	h.flushInterval = d
}
