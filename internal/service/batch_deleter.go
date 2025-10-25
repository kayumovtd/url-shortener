package service

import (
	"context"
	"sync"
	"time"

	"github.com/kayumovtd/url-shortener/internal/logger"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"go.uber.org/zap"
)

type DeleteTask struct {
	UserID   string
	ShortIDs []string
}

type BatchDeleter struct {
	store repository.Store
	log   *logger.Logger

	doneCh       chan struct{}
	inputCh      chan DeleteTask
	aggregatorCh <-chan DeleteTask

	batchSize     int
	flushInterval time.Duration
	timeout       time.Duration
}

func NewBatchDeleter(store repository.Store, log *logger.Logger) *BatchDeleter {
	h := &BatchDeleter{
		store:         store,
		log:           log,
		doneCh:        make(chan struct{}),
		inputCh:       make(chan DeleteTask, 1000),
		batchSize:     100,
		flushInterval: 2 * time.Second,
		timeout:       3 * time.Second,
	}

	workers := h.fanOut(h.inputCh)
	h.aggregatorCh = h.fanIn(workers...)
	go h.runAggregator(h.aggregatorCh)

	return h
}

func (h *BatchDeleter) fanOut(input <-chan DeleteTask) []chan DeleteTask {
	workerCount := 5
	workers := make([]chan DeleteTask, workerCount)
	for i := range workerCount {
		workers[i] = make(chan DeleteTask, 100)
	}

	go func() {
		defer func() {
			for _, ch := range workers {
				close(ch)
			}
		}()

		// Раскидываем таски воркерам по кругу
		i := 0
		for task := range input {
			select {
			case <-h.doneCh:
				return
			case workers[i] <- task:
				i = (i + 1) % workerCount
			}
		}
	}()

	return workers
}

func (h *BatchDeleter) fanIn(inputs ...chan DeleteTask) <-chan DeleteTask {
	finalCh := make(chan DeleteTask)

	var wg sync.WaitGroup

	for _, ch := range inputs {
		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-h.doneCh:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func (h *BatchDeleter) runAggregator(input <-chan DeleteTask) {
	batch := make([]DeleteTask, 0, h.batchSize)

	ticker := time.NewTicker(h.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case task := <-input:
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
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	userURLs := make(map[string][]string)
	for _, t := range batch {
		userURLs[t.UserID] = append(userURLs[t.UserID], t.ShortIDs...)
	}

	for userID, ids := range userURLs {
		if err := h.store.MarkURLsDeleted(ctx, userID, ids); err != nil {
			h.log.Error("failed to mark urls deleted",
				zap.String("userID", userID),
				zap.Strings("ids", ids),
				zap.Error(err),
			)
		}
	}
}

func (h *BatchDeleter) Enqueue(userID string, shortIDs []string) {
	h.inputCh <- DeleteTask{UserID: userID, ShortIDs: shortIDs}
}

func (h *BatchDeleter) Close() {
	close(h.doneCh)
	close(h.inputCh)
}

func (h *BatchDeleter) SetBatchSize(size int) {
	h.batchSize = size
}

func (h *BatchDeleter) SetFlushInterval(d time.Duration) {
	h.flushInterval = d
}
