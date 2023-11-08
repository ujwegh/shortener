package service

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/storage"
)

type (
	ShortenerService interface {
		CreateShortenedURL(originalURL string) (*model.ShortenedURL, error)
		GetShortenedURL(shortURL string) (*model.ShortenedURL, error)
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

func (service *ShortenerServiceImpl) CreateShortenedURL(originalURL string) (*model.ShortenedURL, error) {

	shortURL := generateKey()
	shortenedURL := &model.ShortenedURL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	err := service.storage.WriteShortenedURL(shortenedURL)
	if err != nil {
		return nil, err
	}
	return shortenedURL, nil
}

func (service *ShortenerServiceImpl) GetShortenedURL(shortURL string) (*model.ShortenedURL, error) {
	shortenedURL, err := service.storage.ReadShortenedURL(shortURL)
	if err != nil {
		return nil, err
	}
	return &shortenedURL, nil
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
