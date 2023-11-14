package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/ujwegh/shortener/internal/app/model"
)

type DBStorage struct {
	db *sqlx.DB
}

func NewDBStorage(db *sqlx.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (storage *DBStorage) WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	insertQuery := `INSERT INTO shortened_urls (uuid, short_url, original_url) VALUES ($1, $2, $3);`
	stmt, err := storage.db.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, shortenedURL.UUID, shortenedURL.ShortURL, shortenedURL.OriginalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			getQuery := `SELECT uuid, short_url, original_url, correlation_id FROM shortened_urls su WHERE original_url = $1;`
			err := storage.db.GetContext(ctx, shortenedURL, getQuery, shortenedURL.OriginalURL)
			if err != nil {
				return fmt.Errorf("query existing URL after unique violation: %w", err)
			}
			return tx.Commit()
		}
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback transaction: %w", err)
		}
		return fmt.Errorf("write shortened URL: %w", err)
	}
	return tx.Commit()
}

func (storage *DBStorage) ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	query := `SELECT uuid, short_url, original_url, correlation_id FROM shortened_urls WHERE short_url = $1;`
	shortenedURL := &model.ShortenedURL{}
	err := storage.db.GetContext(ctx, shortenedURL, query, shortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("shortened URL not found")
		}
		return nil, fmt.Errorf("read shortened URL: %w", err)
	}
	return shortenedURL, nil
}

func (storage *DBStorage) Ping(ctx context.Context) error {
	return storage.db.PingContext(ctx)
}

func (storage *DBStorage) WriteBatchShortenedURLSlice(ctx context.Context, urlsSlice []model.ShortenedURL) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	query := `INSERT INTO shortened_urls (uuid, short_url, original_url, correlation_id) 
		VALUES (:uuid, :short_url, :original_url, :correlation_id);`

	toSave := make([]model.ShortenedURL, 0, len(urlsSlice)*20)
	for i, url := range urlsSlice {
		toSave = append(toSave, url)
		if i == len(urlsSlice)-1 || len(toSave) == 20 {
			_, err := storage.db.NamedExecContext(ctx, query, toSave)
			if err != nil {
				if err := tx.Rollback(); err != nil {
					return fmt.Errorf("rollback transaction: %w", err)
				}
				return fmt.Errorf("write batch shortened URL: %w", err)
			}
			toSave = make([]model.ShortenedURL, 0, len(urlsSlice)*20)
		} else if len(toSave) != 20 {
			continue
		}
	}
	return tx.Commit()
}
