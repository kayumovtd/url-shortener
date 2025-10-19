package main

import (
	"log"
	"net/http"

	"github.com/kayumovtd/url-shortener/internal/config"
	"github.com/kayumovtd/url-shortener/internal/handler"
	"github.com/kayumovtd/url-shortener/internal/logger"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"github.com/kayumovtd/url-shortener/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	l, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	defer func() {
		if err := l.Sync(); err != nil {
			log.Printf("logger sync failed: %v", err)
		}
	}()

	store, err := repository.NewStore(cfg, l)
	if err != nil {
		l.Fatal("failed to create store", zap.Error(err))
	}

	svc := service.NewShortenerService(store, cfg.BaseURL)
	defer svc.Close()

	auth := service.NewAuthService(cfg.AuthSecret)
	r := handler.NewRouter(svc, auth, l)

	l.Info("starting server",
		zap.String("address", cfg.Address),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("logLevel", cfg.LogLevel),
		zap.String("fileStoragePath", cfg.FileStoragePath),
		zap.String("databaseDSN", cfg.DatabaseDSN),
	)

	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		l.Fatal("server stopped with error", zap.Error(err))
	}
}
