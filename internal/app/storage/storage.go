package storage

import (
	"context"
	"github.com/ujwegh/shortener/internal/app/model"
)

type Storage interface {
	WriteShortenedURL(shortenedURL *model.ShortenedURL) error
	ReadShortenedURL(shortURL string) (model.ShortenedURL, error)
	Ping(ctx context.Context) error
}
