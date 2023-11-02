package storage

import (
	"encoding/json"
	"github.com/ujwegh/shortener/internal/app/model"
	"os"
)

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
