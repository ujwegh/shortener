package storage

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/model"
	"os"
)

type FileStorage struct {
	filePath string
	urlMap   map[string]model.ShortenedURL
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
	defer func(consumer *Consumer) {
		err := consumer.close()
		if err != nil {
			fmt.Println(err)
		}
	}(consumer)
	var shortenedURLs []model.ShortenedURL
	for {
		shortenedURL, err := consumer.readShortenedURL()
		if err != nil {
			fmt.Println(err)
		}
		if shortenedURL == nil {
			break
		}
		shortenedURLs = append(shortenedURLs, *shortenedURL)
	}
	return shortenedURLs, nil
}

func (fss *FileStorage) WriteShortenedURL(shortenedURL *model.ShortenedURL) error {
	if fss.filePath != "" {
		shortenedURL.UUID = uuid.New()

		producer, err := newProducer(fss.filePath)
		if err != nil {
			return fmt.Errorf("can't create Producer: %w", err)
		}
		defer func(producer *Producer) {
			err := producer.close()
			if err != nil {
				fmt.Println(err)
			}
		}(producer)
		err = producer.writeShortenedURL(shortenedURL)
		if err != nil {
			return fmt.Errorf("can't write shortened URL: %w", err)
		}
	}
	fss.urlMap[shortenedURL.ShortURL] = *shortenedURL
	return nil
}

func (fss *FileStorage) ReadShortenedURL(shortURL string) (model.ShortenedURL, error) {
	shortenedURL := fss.urlMap[shortURL]
	return shortenedURL, nil
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func newProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) writeShortenedURL(shortenedURL *model.ShortenedURL) error {
	return p.encoder.Encode(&shortenedURL)
}

func (p *Producer) close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) readShortenedURL() (*model.ShortenedURL, error) {
	url := &model.ShortenedURL{}
	if err := c.decoder.Decode(&url); err != nil {
		return nil, err
	}

	return url, nil
}

func (c *Consumer) close() error {
	return c.file.Close()
}
