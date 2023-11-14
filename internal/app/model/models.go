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
	}
)
