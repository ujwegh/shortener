package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/storage"
)

type (
	ShortenerService interface {
		CreateShortenedURL(ctx context.Context, originalURL string) (*model.ShortenedURL, error)
		GetShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error)
		BatchCreateShortenedURLs(ctx context.Context, dtos []model.ShortenedURL) (*[]model.ShortenedURL, error)
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

func (ss *ShortenerServiceImpl) CreateShortenedURL(ctx context.Context, originalURL string) (*model.ShortenedURL, error) {

	shortURL := generateKey()
	shortenedURL := &model.ShortenedURL{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	err := ss.storage.WriteShortenedURL(ctx, shortenedURL)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	return shortenedURL, nil
}

func (ss *ShortenerServiceImpl) GetShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	shortenedURL, err := ss.storage.ReadShortenedURL(ctx, shortURL)
	if err != nil {
		println("error: %v", err)
		fmt.Printf("error: %v", err)
		return nil, err
	}
	return shortenedURL, nil
}

func (ss *ShortenerServiceImpl) BatchCreateShortenedURLs(ctx context.Context, urls []model.ShortenedURL) (*[]model.ShortenedURL, error) {
	for i := range urls {
		urls[i].UUID = uuid.New()
		urls[i].ShortURL = generateKey()
	}
	err := ss.storage.WriteBatchShortenedURLSlice(ctx, urls)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	return &urls, nil
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
