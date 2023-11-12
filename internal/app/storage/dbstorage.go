package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ujwegh/shortener/internal/app/model"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (storage *DBStorage) WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	query := `INSERT INTO shortened_urls (uuid, short_url, original_url) VALUES ($1, $2, $3);`
	_, err := storage.db.ExecContext(ctx, query, shortenedURL.UUID, shortenedURL.ShortURL, shortenedURL.OriginalURL)
	if err != nil {
		return fmt.Errorf("write shortened URL: %w", err)
	}
	return nil
}

func (storage *DBStorage) ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	query := `SELECT uuid, short_url, original_url FROM shortened_urls WHERE short_url = $1;`
	row := storage.db.QueryRowContext(ctx, query, shortURL)

	var shortenedURL model.ShortenedURL
	err := row.Scan(&shortenedURL.UUID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("shortened URL not found")
		}
		return nil, fmt.Errorf("read shortened URL: %w", err)
	}
	return &shortenedURL, nil
}

func (storage *DBStorage) Ping(ctx context.Context) error {
	return storage.db.PingContext(ctx)
}
