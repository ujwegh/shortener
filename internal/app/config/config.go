package config

import (
	"flag"
)

type AppConfig struct {
	ServerAddr       string
	ShortenedURLAddr string
}

func ParseFlags() AppConfig {
	config := AppConfig{}
	flag.StringVar(&config.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.ShortenedURLAddr, "b", "http://localhost:8080", "address and port for shortened url")
	flag.Parse()
	return config
}
