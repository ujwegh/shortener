package config

import (
	"flag"
	"os"
)

type AppConfig struct {
	ServerAddr            string
	ShortenedURLAddr      string
	ShortenedURLsFilePath string
	UserURLsFilePath      string
	LogLevel              string
	DatabaseDSN           string
	ContextTimeoutSec     int
	TokenSecretKey        string
}

func ParseFlags() AppConfig {
	// Define defaults
	const (
		defaultServerAddress         = "localhost:8080"
		defaultShortenedURLAddress   = "http://localhost:8080"
		defaultShortenedURLsFilePath = "/tmp/short-url-db.json"
		defaultUserURLsPath          = "/tmp/user-url-db.json"
		defaultLogLevel              = "info"
		defaultDatabaseDSN           = "" //postgres://postgres:mysecretpassword@localhost:5432/postgres
		defaultContextTimeoutSec     = 5
	)

	// Initialize AppConfig with defaults
	config := AppConfig{
		ServerAddr:            defaultServerAddress,
		ShortenedURLAddr:      defaultShortenedURLAddress,
		ShortenedURLsFilePath: defaultShortenedURLsFilePath,
		UserURLsFilePath:      defaultUserURLsPath,
		LogLevel:              defaultLogLevel,
		DatabaseDSN:           defaultDatabaseDSN,
		ContextTimeoutSec:     defaultContextTimeoutSec,
	}

	// Set flags
	flag.StringVar(&config.ServerAddr, "a", config.ServerAddr, "address and port to run server")
	flag.StringVar(&config.ShortenedURLAddr, "b", config.ShortenedURLAddr, "address and port for shortened url")
	flag.StringVar(&config.LogLevel, "ll", config.LogLevel, "logging level")
	flag.StringVar(&config.ShortenedURLsFilePath, "f", config.ShortenedURLsFilePath, "file storage path")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database dsn")
	flag.Parse()

	// Override with environment variables if they exist
	if envVal := os.Getenv("SERVER_ADDRESS"); envVal != "" {
		config.ServerAddr = envVal
	}
	if envVal := os.Getenv("BASE_URL"); envVal != "" {
		config.ShortenedURLAddr = envVal
	}
	if envVal := os.Getenv("LOG_LEVEL"); envVal != "" {
		config.LogLevel = envVal
	}
	if envVal := os.Getenv("FILE_STORAGE_PATH"); envVal != "" {
		config.ShortenedURLsFilePath = envVal
	}
	if envVal := os.Getenv("DATABASE_DSN"); envVal != "" {
		config.DatabaseDSN = envVal
	}

	return config
}
