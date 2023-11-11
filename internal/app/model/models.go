package model

import "github.com/google/uuid"

type (
	//easyjson:json
	ShortenedURL struct {
		UUID        uuid.UUID `json:"uuid"`
		ShortURL    string    `json:"short_url"`
		OriginalURL string    `json:"original_url"`
	}
)
