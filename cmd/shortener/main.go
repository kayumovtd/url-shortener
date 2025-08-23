package main

import (
	"log"
	"net/http"

	"github.com/kayumovtd/url-shortener/internal/config"
	"github.com/kayumovtd/url-shortener/internal/handler"
	"github.com/kayumovtd/url-shortener/internal/repository"
)

func main() {
	cfg := config.NewConfig()
	store := repository.NewInMemoryStore()
	r := handler.NewRouter(store, cfg.BaseURL)

	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
