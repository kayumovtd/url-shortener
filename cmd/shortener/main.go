package main

import (
	"log"
	"net/http"

	"github.com/kayumovtd/url-shortener/internal/config"
	"github.com/kayumovtd/url-shortener/internal/handler"
	"github.com/kayumovtd/url-shortener/internal/logger"
	"github.com/kayumovtd/url-shortener/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	store := repository.NewInMemoryStore()
	r := handler.NewRouter(store, cfg.BaseURL)

	logger.Log.Info("starting server",
		zap.String("address", cfg.Address),
		zap.String("baseURL", cfg.BaseURL),
		zap.String("logLevel", cfg.LogLevel),
	)

	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
