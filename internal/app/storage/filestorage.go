package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/model"
	"io"
	"sync"
)

type FileStorage struct {
	shortenedURLsFilePath string
	userURLsFilePath      string
	shortURLMap           map[string]model.ShortenedURL    // shortURL -> ShortenedURL
	uuidURLMap            map[uuid.UUID]model.ShortenedURL // uuid -> ShortenedURL
	mutex                 sync.Mutex
}

func (fs *FileStorage) DeleteUserURLs(ctx context.Context, userURL *uuid.UUID, shortURLKeys []string) error {
	//TODO implement me
	panic("implement me")
}

func NewFileStorage(cfg config.AppConfig) *FileStorage {
	urlMap := make(map[string]model.ShortenedURL)
	uuidMap := make(map[uuid.UUID]model.ShortenedURL)
	storage := FileStorage{
		shortenedURLsFilePath: cfg.ShortenedURLsFilePath,
		userURLsFilePath:      cfg.UserURLsFilePath,
	}
	if cfg.ShortenedURLsFilePath != "" {
		ls, err := storage.readAllShortenedURLs()
		if err != nil {
			panic(err)
		}
		for _, l := range ls {
			urlMap[l.ShortURL] = l
			uuidMap[l.UUID] = l
		}
	}
	storage.shortURLMap = urlMap
	storage.uuidURLMap = uuidMap
	return &storage
}

func (fs *FileStorage) ReadUserURLs(ctx context.Context, uid *uuid.UUID) ([]model.ShortenedURL, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	consumer, err := newConsumer(fs.userURLsFilePath)
	if err != nil {
		return nil, fmt.Errorf("can't create Consumer: %w", err)
	}
	defer consumer.close()

	var shortenedURLs []model.ShortenedURL
	for {
		userURL := &model.UserURL{}
		err := consumer.readObject(userURL)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if userURL.UUID == *uid {
			shortenedURLs = append(shortenedURLs, fs.uuidURLMap[userURL.ShortenedURLUUID])
		}
	}
	return shortenedURLs, nil
}

func (fs *FileStorage) CreateUserURL(ctx context.Context, userURL *model.UserURL) error {

	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	if fs.userURLsFilePath != "" {
		producer, err := newProducer(fs.userURLsFilePath)
		if err != nil {
			return fmt.Errorf("can't create Producer: %w", err)
		}
		defer producer.close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = producer.writeObject(userURL)
			if err != nil {
				return fmt.Errorf("can't write user URL: %w", err)
			}
		}
	}
	return nil
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return fmt.Errorf("file storage doesn't support Ping() method")
}

func (fs *FileStorage) readAllShortenedURLs() ([]model.ShortenedURL, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	consumer, err := newConsumer(fs.shortenedURLsFilePath)
	if err != nil {
		return nil, fmt.Errorf("can't create Consumer: %w", err)
	}
	defer consumer.close()

	var shortenedURLs []model.ShortenedURL
	for {
		shortenedURL := &model.ShortenedURL{}
		err := consumer.readObject(shortenedURL)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		shortenedURLs = append(shortenedURLs, *shortenedURL)
	}
	return shortenedURLs, nil
}

func (fs *FileStorage) WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	if fs.shortenedURLsFilePath != "" {
		producer, err := newProducer(fs.shortenedURLsFilePath)
		if err != nil {
			return fmt.Errorf("can't create Producer: %w", err)
		}
		defer producer.close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = producer.writeObject(shortenedURL)
			if err != nil {
				return fmt.Errorf("can't write shortened URL: %w", err)
			}
		}
	}
	fs.shortURLMap[shortenedURL.ShortURL] = *shortenedURL
	return nil
}

func (fs *FileStorage) ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		shortenedURL := fs.shortURLMap[shortURL]
		return &shortenedURL, nil
	}
}

func (fs *FileStorage) WriteBatchShortenedURLSlice(ctx context.Context, slice []model.ShortenedURL) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if fs.shortenedURLsFilePath != "" {
		producer, err := newProducer(fs.shortenedURLsFilePath)
		if err != nil {
			return fmt.Errorf("can't create Producer: %w", err)
		}
		defer producer.close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			for _, shortenedURL := range slice {
				err = producer.writeObject(shortenedURL)
				if err != nil {
					return fmt.Errorf("can't write shortened URL: %w", err)
				}
			}
		}
	}
	for _, shortenedURL := range slice {
		fs.shortURLMap[shortenedURL.ShortURL] = shortenedURL
	}
	return nil
}
