package model

import "github.com/google/uuid"

type (
	ShortenRequestDto struct {
		URL string `json:"url"`
	}
	ShortenResponseDto struct {
		Result string `json:"result"`
	}
	ShortenedURL struct {
		UUID        uuid.UUID `json:"uuid"`
		ShortURL    string    `json:"short_url"`
		OriginalURL string    `json:"original_url"`
	}
)
