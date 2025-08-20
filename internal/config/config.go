package config

import (
	"flag"
)

type Config struct {
	Address string
	BaseURL string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(
		&cfg.Address,
		"a",
		":8080",
		"HTTP server listen address",
	)

	flag.StringVar(
		&cfg.BaseURL,
		"b",
		"http://localhost:8080",
		"Base URL for shortened URLs",
	)

	flag.Parse()

	return cfg
}
