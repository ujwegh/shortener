package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/storage"
	"go.uber.org/zap"
	"time"
)

type (
	ShortenerService interface {
		CreateShortenedURL(ctx context.Context, userUID *uuid.UUID, originalURL string) (*model.ShortenedURL, error)
		GetShortenedURL(ctx context.Context, url string) (*model.ShortenedURL, error)
		BatchCreateShortenedURLs(ctx context.Context, dtos []model.ShortenedURL) (*[]model.ShortenedURL, error)
		GetUserShortenedURLs(ctx context.Context, userUID *uuid.UUID) (*[]model.ShortenedURL, error)
		DeleteUserShortenedURLs(ctx context.Context, userUID *uuid.UUID, shortURLKeys []string) error
	}
	ShortenerServiceImpl struct {
		storage     storage.Storage
		taskChannel chan Task
	}
	Task struct {
		UserUID      uuid.UUID
		ShortURLKeys []string
	}
)

func NewShortenerService(storage storage.Storage, taskChannel chan Task) *ShortenerServiceImpl {
	return &ShortenerServiceImpl{
		storage:     storage,
		taskChannel: taskChannel,
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

func (ss *ShortenerServiceImpl) DeleteUserShortenedURLs(ctx context.Context, userUID *uuid.UUID, shortURLKeys []string) error {
	const chunkSize = 20
	slice := make([]string, 0, chunkSize)
	for _, shortURL := range shortURLKeys {
		slice = append(slice, shortURL)
		if len(slice) == chunkSize {
			ss.taskChannel <- Task{
				UserUID:      *userUID,
				ShortURLKeys: slice,
			}
			slice = nil
		}
	}
	if len(slice) > 0 {
		ss.taskChannel <- Task{
			UserUID:      *userUID,
			ShortURLKeys: slice,
		}
	}
	return nil
}

func (ss *ShortenerServiceImpl) BatchProcess(ctx context.Context, taskChannel <-chan Task) {
	buffer := make(map[uuid.UUID][]string)
	taskCount := 0

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case task, ok := <-taskChannel:
			if !ok {
				deleteUserURLs(ss, buffer)
				return
			}
			buffer[task.UserUID] = append(buffer[task.UserUID], task.ShortURLKeys...)
			taskCount++

			if taskCount >= 20 {
				deleteUserURLs(ss, buffer)
				buffer = make(map[uuid.UUID][]string)
				taskCount = 0
			}
		case <-ticker.C:
			// Periodically flush the map
			if taskCount > 0 {
				deleteUserURLs(ss, buffer)
				buffer = make(map[uuid.UUID][]string)
				taskCount = 0
			}

		case <-ctx.Done():
			deleteUserURLs(ss, buffer)
			return
		}
	}
}

func deleteUserURLs(ss *ShortenerServiceImpl, buffer map[uuid.UUID][]string) {
	err := ss.storage.DeleteBulk(context.Background(), buffer)
	if err != nil {
		logger.Log.Error("failed to delete bulk user URLs", zap.Error(err))
	}
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
