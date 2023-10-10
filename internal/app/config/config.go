package config

import (
	"flag"
)

type AppConfig struct {
	ServerAddr       string `env:"SERVER_ADDRESS"`
	ShortenedURLAddr string `env:"BASE_URL"`
}

func ParseFlags() AppConfig {
	defaultServerAddress := "localhost:8080"
	defaultShortenedURLAddress := "http://localhost:8080"

	config := AppConfig{}
	if config.ServerAddr == "" {
		flag.StringVar(&config.ServerAddr, "a", defaultServerAddress, "address and port to run server")
	}
	if config.ShortenedURLAddr == "" {
		flag.StringVar(&config.ShortenedURLAddr, "b", defaultShortenedURLAddress, "address and port for shortened url")
	}
	flag.Parse()
	return config
}
