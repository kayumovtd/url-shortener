package repository

import (
	"fmt"

	"github.com/kayumovtd/url-shortener/internal/config"
	"github.com/kayumovtd/url-shortener/internal/logger"
)

func NewStore(cfg *config.Config, l *logger.Logger) (Store, error) {
	if cfg.DatabaseDSN != "" {
		dbStore, err := NewDBStore(cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to init db store: %w", err)
		}
		l.Info("using db store")
		return dbStore, nil
	}

	if cfg.FileStoragePath != "" {
		fileStore, err := NewFileStore(cfg.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("failed to init file store: %w", err)
		}
		l.Info("using file store")
		return fileStore, nil
	}

	l.Info("using in-memory store")
	return NewInMemoryStore(), nil
}
