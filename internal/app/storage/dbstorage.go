package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/ujwegh/shortener/internal/app/config"
	appErrors "github.com/ujwegh/shortener/internal/app/errors"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/migrations"
)

type DBStorage struct {
	db *sqlx.DB
}

func (storage *DBStorage) ReadUserURLs(ctx context.Context, uid *uuid.UUID) ([]model.ShortenedURL, error) {
	query := `SELECT su.uuid, su.short_url, su.original_url, su.correlation_id, su.is_deleted
	FROM shortened_urls su
	JOIN user_urls uu on su.uuid = uu.shortened_url_uuid
	WHERE uu.uuid = $1;`
	shortenedURLs := make([]model.ShortenedURL, 0)
	err := storage.db.SelectContext(ctx, &shortenedURLs, query, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return shortenedURLs, nil
		}
		return nil, fmt.Errorf("read user URLs: %w", err)
	}
	return shortenedURLs, nil
}

func NewDBStorage(cfg config.AppConfig) *DBStorage {
	db := Open(cfg.DatabaseDSN)
	// Migrate the database
	err := MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	return &DBStorage{db: db}
}

func (storage *DBStorage) WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	insertQuery := `INSERT INTO shortened_urls (uuid, short_url, original_url) VALUES ($1, $2, $3);`
	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, shortenedURL.UUID, shortenedURL.ShortURL, shortenedURL.OriginalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return appErrors.New(err, "unique violation")
		}
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback transaction: %w", err)
		}
		return fmt.Errorf("write shortened URL: %w", err)
	}
	return tx.Commit()
}

func (storage *DBStorage) ReadShortenedURL(ctx context.Context, url string) (*model.ShortenedURL, error) {
	query := `SELECT uuid, short_url, original_url, correlation_id, is_deleted
	FROM shortened_urls WHERE short_url = $1 or original_url = $1;`
	shortenedURL := &model.ShortenedURL{}
	err := storage.db.GetContext(ctx, shortenedURL, query, url)
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

func (storage *DBStorage) CreateUserURL(ctx context.Context, userURL *model.UserURL) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	query := `INSERT INTO user_urls (uuid, shortened_url_uuid) VALUES ($1, $2);`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userURL.UUID, userURL.ShortenedURLUUID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback transaction: %w", err)
		}
		return fmt.Errorf("write user URL: %w", err)
	}
	return tx.Commit()
}

func (storage *DBStorage) DeleteUserURLs(ctx context.Context, userURL *uuid.UUID, shortURLKeys []string) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	query := `update shortened_urls set is_deleted = true 
              from user_urls uu
              where shortened_urls.uuid = uu.shortened_url_uuid
    		  AND uu.uuid = $1
    		  AND shortened_urls.short_url = any ($2)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userURL, shortURLKeys)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback transaction: %w", err)
		}
		return fmt.Errorf("delete user URLs: %w", err)
	}
	return tx.Commit()
}
