package handlers

import (
	"database/sql"
	"fmt"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"time"
)

type (
	ShortenerHandlers struct {
		shortenerService service.ShortenerService
		shortenedURLAddr string
		storage          storage.Storage
		contextTimeout   time.Duration
	}
	//easyjson:json
	ShortenRequestDto struct {
		URL string `json:"url"`
	}
	//easyjson:json
	ShortenResponseDto struct {
		Result string `json:"result"`
	}
	//easyjson:json
	ExternalShortenedURLRequestDto struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	//easyjson:json
	ExternalShortenedURLResponseDto struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	//easyjson:json
	ExternalShortenedURLRequestDtoSlice []ExternalShortenedURLRequestDto
	//easyjson:json
	ExternalShortenedURLResponseDtoSlice []ExternalShortenedURLResponseDto
	//easyjson:json
	UserURLDto struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	//easyjson:json
	UserURLDtoSlice []UserURLDto

	//easyjson:json
	DeleteUserURLsDto []string
)

func mapShortenedURLToExternalResponse(sh *ShortenerHandlers, slice []model.ShortenedURL) ExternalShortenedURLResponseDtoSlice {
	var responseSlice []ExternalShortenedURLResponseDto
	for _, item := range slice {
		responseItem := ExternalShortenedURLResponseDto{
			CorrelationID: item.CorrelationID.String,
			ShortURL:      fmt.Sprintf("%s/%s", sh.shortenedURLAddr, item.ShortURL),
		}
		responseSlice = append(responseSlice, responseItem)
	}
	return responseSlice
}
func mapShortenedURLToUserURLDtoSlice(sh *ShortenerHandlers, slice []model.ShortenedURL) UserURLDtoSlice {
	var responseSlice []UserURLDto
	for _, item := range slice {
		responseItem := UserURLDto{
			OriginalURL: item.OriginalURL,
			ShortURL:    fmt.Sprintf("%s/%s", sh.shortenedURLAddr, item.ShortURL),
		}
		responseSlice = append(responseSlice, responseItem)
	}
	return responseSlice
}

func mapExternalRequestToShortenedURL(slice []ExternalShortenedURLRequestDto) *[]model.ShortenedURL {
	var shortenedURLs []model.ShortenedURL
	for _, item := range slice {
		shortenedURL := model.ShortenedURL{
			CorrelationID: sql.NullString{
				String: item.CorrelationID,
				Valid:  true,
			},
			OriginalURL: item.OriginalURL,
		}
		shortenedURLs = append(shortenedURLs, shortenedURL)
	}
	return &shortenedURLs
}
