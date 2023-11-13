package model

import "github.com/google/uuid"

type (
	//easyjson:json
	ShortenedURL struct {
		UUID          uuid.UUID `json:"uuid" db:"uuid"`
		ShortURL      string    `json:"short_url" db:"short_url"`
		OriginalURL   string    `json:"original_url" db:"original_url"`
		CorrelationID string    `json:"correlation_id" db:"correlation_id"`
	}
)
