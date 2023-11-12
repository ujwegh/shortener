package storage

import (
	"context"
	"github.com/ujwegh/shortener/internal/app/model"
)

type Storage interface {
	WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error
	ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error)
	Ping(ctx context.Context) error
}
