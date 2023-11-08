package storage

import (
	"context"
	"database/sql"
	"github.com/ujwegh/shortener/internal/app/model"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (db *DBStorage) WriteShortenedURL(shortenedURL *model.ShortenedURL) error {
	//TODO implement me
	panic("implement me")
}

func (db *DBStorage) ReadShortenedURL(shortURL string) (model.ShortenedURL, error) {
	//TODO implement me
	panic("implement me")
}

func (db *DBStorage) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
