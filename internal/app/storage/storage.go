package storage

import (
	"context"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/model"
)

type Storage interface {
	WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error
	ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error)
	Ping(ctx context.Context) error
	WriteBatchShortenedURLSlice(ctx context.Context, slice []model.ShortenedURL) error
}

func NewStorage(cfg config.AppConfig) Storage {
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("Using database storage.")
		return NewDBStorage(cfg)
	} else {
		logger.Log.Info("Using in-memory storage.")
		return NewFileStorage(cfg.FileStoragePath)
	}
}
