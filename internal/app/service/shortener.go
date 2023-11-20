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
		CreateShortenedURL(ctx context.Context, userUID *uuid.UUID, originalURL string) (*model.ShortenedURL, error)
		GetShortenedURL(ctx context.Context, url string) (*model.ShortenedURL, error)
		BatchCreateShortenedURLs(ctx context.Context, dtos []model.ShortenedURL) (*[]model.ShortenedURL, error)
		GetUserShortenedURLs(ctx context.Context, userUID *uuid.UUID) (*[]model.ShortenedURL, error)
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

func (ss *ShortenerServiceImpl) CreateShortenedURL(ctx context.Context, userUID *uuid.UUID, originalURL string) (*model.ShortenedURL, error) {

	shortURL := generateKey()
	shortenedURL := &model.ShortenedURL{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	err := ss.storage.WriteShortenedURL(ctx, shortenedURL)
	if err != nil {
		return nil, err
	}
	userURL := &model.UserURL{
		UUID:             *userUID,
		ShortenedURLUUID: shortenedURL.UUID,
	}
	err = ss.storage.CreateUserURL(ctx, userURL)
	if err != nil {
		return nil, err
	}
	return shortenedURL, nil
}

func (ss *ShortenerServiceImpl) GetShortenedURL(ctx context.Context, url string) (*model.ShortenedURL, error) {
	shortenedURL, err := ss.storage.ReadShortenedURL(ctx, url)
	if err != nil {
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
		return nil, err
	}
	return &urls, nil
}

func (ss *ShortenerServiceImpl) GetUserShortenedURLs(ctx context.Context, userUID *uuid.UUID) (*[]model.ShortenedURL, error) {
	userURLs, err := ss.storage.ReadUserURLs(ctx, userUID)
	if err != nil {
		return nil, err
	}
	return &userURLs, nil
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
