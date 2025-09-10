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

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	store, err := repository.NewFileStore(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("failed to create file store", zap.Error(err))
	}

	svc := service.NewShortenerService(store, cfg.BaseURL)
	r := handler.NewRouter(svc)

	logger.Log.Info("starting server",
		zap.String("address", cfg.Address),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("logLevel", cfg.LogLevel),
		zap.String("fileStoragePath", cfg.FileStoragePath),
	)

	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
