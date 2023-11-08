package storage

import (
	"encoding/json"
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

func (p *Producer) writeObject(obj interface{}) error {
	if err := p.encoder.Encode(obj); err != nil {
		return err
	}
	return nil
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

func (c *Consumer) readObject(obj interface{}) error {
	if err := c.decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

func (c *Consumer) close() error {
	return c.file.Close()
}
