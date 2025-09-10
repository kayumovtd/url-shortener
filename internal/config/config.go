package config

import (
	"flag"
	"os"
)

const (
	defaultAddress  = ":8080"
	defaultBaseURL  = "http://localhost:8080"
	defaultLogLevel = "info"
	envServerAddr   = "SERVER_ADDRESS"
	envBaseURL      = "BASE_URL"
)

type Config struct {
	Address  string
	BaseURL  string
	LogLevel string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", defaultAddress, "HTTP server listen address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened URLs")
	flag.StringVar(&cfg.LogLevel, "l", defaultLogLevel, "Level for logs")
	flag.Parse()

	if addr := os.Getenv(envServerAddr); addr != "" {
		cfg.Address = addr
	}
	if base := os.Getenv(envBaseURL); base != "" {
		cfg.BaseURL = base
	}

	return cfg
}
