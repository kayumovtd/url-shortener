package config

import (
	"flag"
	"os"
)

const (
	defaultAddress         = ":8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultLogLevel        = "info"
	defaultFileStoragePath = "storage.json"
	defaultDatabaseDSN     = ""

	envServerAddr      = "SERVER_ADDRESS"
	envBaseURL         = "BASE_URL"
	envFileStoragePath = "FILE_STORAGE_PATH"
	envDatabaseDSN     = "DATABASE_DSN"
)

type Config struct {
	Address         string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", defaultAddress, "HTTP server listen address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened URLs")
	flag.StringVar(&cfg.LogLevel, "l", defaultLogLevel, "Level for logs")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", defaultDatabaseDSN, "PostgreSQL DSN")
	flag.Parse()

	if v, ok := os.LookupEnv(envServerAddr); ok {
		cfg.Address = v
	}
	if v, ok := os.LookupEnv(envBaseURL); ok {
		cfg.BaseURL = v
	}
	if v, ok := os.LookupEnv(envFileStoragePath); ok {
		cfg.FileStoragePath = v
	}
	if v, ok := os.LookupEnv(envDatabaseDSN); ok {
		cfg.DatabaseDSN = v
	}

	return cfg
}
