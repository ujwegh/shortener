package config

import (
	"flag"
	"os"
)

type AppConfig struct {
	ServerAddr       string `env:"SERVER_ADDRESS"`
	ShortenedURLAddr string `env:"BASE_URL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	LogLevel         string
	DatabaseDSN      string
}

func ParseFlags() AppConfig {
	defaultServerAddress := "localhost:8080"
	defaultShortenedURLAddress := "http://localhost:8080"
	defaultFileStoragePath := "/tmp/short-url-db.json"
	defaultLogLevel := "info"
	defaultDatabaseDSN := ""

	config := AppConfig{}
	fileStoragePath, fileStoragePathExist := os.LookupEnv("FILE_STORAGE_PATH")
	config.FileStoragePath = fileStoragePath
	if !fileStoragePathExist {
		flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "file storage path")
	}
	if config.ServerAddr == "" {
		flag.StringVar(&config.ServerAddr, "a", defaultServerAddress, "address and port to run server")
	}
	if config.ShortenedURLAddr == "" {
		flag.StringVar(&config.ShortenedURLAddr, "b", defaultShortenedURLAddress, "address and port for shortened url")
	}
	flag.StringVar(&config.LogLevel, "ll", defaultLogLevel, "logging level")

	databaseDSN, databaseDSNExist := os.LookupEnv("DATABASE_DSN")
	config.DatabaseDSN = databaseDSN
	if !databaseDSNExist {
		flag.StringVar(&config.DatabaseDSN, "d", defaultDatabaseDSN, "database dsn")
	}

	flag.Parse()
	return config
}
