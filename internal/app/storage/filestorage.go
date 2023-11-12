package storage

import (
	"context"
	"fmt"
	"github.com/ujwegh/shortener/internal/app/model"
	"io"
)

type FileStorage struct {
	filePath string
	urlMap   map[string]model.ShortenedURL
}

func (fss *FileStorage) Ping(ctx context.Context) error {
	return fmt.Errorf("file storage doesn't support Ping() method")
}

func NewFileStorage(filePath string) *FileStorage {
	urlMap := make(map[string]model.ShortenedURL)
	storage := FileStorage{
		filePath: filePath,
	}
	if filePath != "" {
		ls, err := storage.readAllShortenedURLs()
		if err != nil {
			panic(err)
		}
		for _, l := range ls {
			urlMap[l.ShortURL] = l
		}
	}
	storage.urlMap = urlMap
	return &storage
}

func (fss *FileStorage) readAllShortenedURLs() ([]model.ShortenedURL, error) {
	consumer, err := newConsumer(fss.filePath)
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

func (fss *FileStorage) WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	if fss.filePath != "" {
		producer, err := newProducer(fss.filePath)
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
	fss.urlMap[shortenedURL.ShortURL] = *shortenedURL
	return nil
}

func (fss *FileStorage) ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		shortenedURL := fss.urlMap[shortURL]
		return &shortenedURL, nil
	}
}
