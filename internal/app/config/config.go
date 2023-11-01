package config

import (
	"flag"
	"github.com/ujwegh/shortener/internal/app/middlware"
)

type AppConfig struct {
	ServerAddr       string `env:"SERVER_ADDRESS"`
	ShortenedURLAddr string `env:"BASE_URL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
}

func ParseFlags() AppConfig {
	defaultServerAddress := "localhost:8080"
	defaultShortenedURLAddress := "http://localhost:8080"
	defaultFileStoragePath := "/tmp/short-url-db.json"

	config := AppConfig{}
	if config.ServerAddr == "" {
		flag.StringVar(&config.ServerAddr, "a", defaultServerAddress, "address and port to run server")
	}
	if config.ShortenedURLAddr == "" {
		flag.StringVar(&config.ShortenedURLAddr, "b", defaultShortenedURLAddress, "address and port for shortened url")
	}
	if config.FileStoragePath == "" {
		flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "file storage path")
	}
	initLogger()
	flag.Parse()
	return config
}

func initLogger() {
	defaultLogLevel := "info"
	var logLevel = ""
	flag.StringVar(&logLevel, "ll", defaultLogLevel, "logging level")
	err := middlware.LoggerInitialize(logLevel)
	if err != nil {
		panic(err)
	}
}
