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
	tx, err := storage.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	query := `INSERT INTO shortened_urls (uuid, short_url, original_url) VALUES ($1, $2, $3);`
	_, err = storage.db.ExecContext(ctx, query, shortenedURL.UUID, shortenedURL.ShortURL, shortenedURL.OriginalURL)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback transaction: %w", err)
		}
		return fmt.Errorf("write shortened URL: %w", err)
	}
	return tx.Commit()
}

func (storage *DBStorage) ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	query := `SELECT uuid, short_url, original_url FROM shortened_urls WHERE short_url = $1;`
	stmt, err := storage.db.PrepareContext(ctx, query)
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, shortURL)
	var shortenedURL model.ShortenedURL
	err = row.Scan(&shortenedURL.UUID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL)
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
