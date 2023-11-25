package model

import (
	"database/sql"
	"github.com/google/uuid"
)

type (
	//easyjson:json
	ShortenedURL struct {
		UUID          uuid.UUID      `json:"uuid" db:"uuid"`
		ShortURL      string         `json:"short_url" db:"short_url"`
		OriginalURL   string         `json:"original_url" db:"original_url"`
		CorrelationID sql.NullString `json:"correlation_id" db:"correlation_id"`
		DeletedFlag   bool           `json:"is_deleted" db:"is_deleted"`
	}
	//easyjson:json
	UserURL struct {
		UUID             uuid.UUID `json:"uuid" db:"uuid"`
		ShortenedURLUUID uuid.UUID `json:"shortened_url_uuid" db:"shortened_url_uuid"`
	}
)
