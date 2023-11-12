package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/storage"
)

type (
	ShortenerService interface {
		CreateShortenedURL(ctx context.Context, originalURL string) (*model.ShortenedURL, error)
		GetShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error)
	}
	ShortenerServiceImpl struct {
		storage storage.Storage
	}
)

func NewShortenerService(storage storage.Storage) *ShortenerServiceImpl {
	return &ShortenerServiceImpl{
		storage: storage,
	}
}

func (service *ShortenerServiceImpl) CreateShortenedURL(ctx context.Context, originalURL string) (*model.ShortenedURL, error) {

	shortURL := generateKey()
	shortenedURL := &model.ShortenedURL{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	err := service.storage.WriteShortenedURL(ctx, shortenedURL)
	if err != nil {
		return nil, err
	}
	return shortenedURL, nil
}

func (service *ShortenerServiceImpl) GetShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	shortenedURL, err := service.storage.ReadShortenedURL(ctx, shortURL)
	if err != nil {
		return nil, err
	}
	return shortenedURL, nil
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
