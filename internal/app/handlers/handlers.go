package handlers

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"io"
	"net/http"
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
)

func NewShortenerHandlers(shortenedURLAddr string, contextTimeout int, service service.ShortenerService, storage storage.Storage) *ShortenerHandlers {
	return &ShortenerHandlers{
		shortenerService: service,
		storage:          storage,
		shortenedURLAddr: shortenedURLAddr,
		contextTimeout:   time.Duration(contextTimeout) * time.Second,
	}
}

func (sh *ShortenerHandlers) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "Url is empty", http.StatusBadRequest)
		return
	}
	shortenedURL, err := sh.shortenerService.CreateShortenedURL(ctx, originalURL)
	if err != nil {
		http.Error(w, "Unable to create shortened URL", http.StatusInternalServerError)
		return
	}
	if contextHasError(w, ctx) {
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", sh.shortenedURLAddr, shortenedURL.ShortURL)
}

func (sh *ShortenerHandlers) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	request := ShortenRequestDto{}
	err = request.UnmarshalJSON(body)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	if request.URL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}
	shortenedURL, err := sh.shortenerService.CreateShortenedURL(ctx, request.URL)
	if err != nil {
		http.Error(w, "Unable to create shortened URL", http.StatusInternalServerError)
		return
	}
	response := &ShortenResponseDto{Result: fmt.Sprintf("%s/%s", sh.shortenedURLAddr, shortenedURL.ShortURL)}
	rawBytes, err := response.MarshalJSON()
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	if contextHasError(w, ctx) {
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", rawBytes)
}

func (sh *ShortenerHandlers) HandleShortenedURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	shortKey := chi.URLParam(r, "id")
	shortenedURL, err := sh.shortenerService.GetShortenedURL(ctx, shortKey)
	if err != nil {
		http.Error(w, "Unable to get shortened URL", http.StatusInternalServerError)
		return
	}
	originalURL := shortenedURL.OriginalURL
	if originalURL == "" {
		http.Error(w, "Shortened url not found", http.StatusNotFound)
		return
	}

	if contextHasError(w, ctx) {
		return
	}
	w.Header().Add("Location", originalURL)
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (sh *ShortenerHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	err := sh.storage.Ping(ctx)
	if err != nil {
		http.Error(w, "Unable to ping storage", http.StatusInternalServerError)
		return
	}

	if contextHasError(w, ctx) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (sh *ShortenerHandlers) APIShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	request := ExternalShortenedURLRequestDtoSlice{}
	err = request.UnmarshalJSON(body)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	var dtos []ExternalShortenedURLRequestDto = request
	if len(dtos) == 0 {
		http.Error(w, "Batch is empty", http.StatusBadRequest)
		return
	}
	urls := mapExternalRequestToShortenedURL(dtos)
	shortenedURLs, err := sh.shortenerService.BatchCreateShortenedURLs(ctx, *urls)
	if err != nil {
		http.Error(w, "Unable to batch insert shortened URLs", http.StatusInternalServerError)
		return
	}
	response := mapShortenedURLToExternalResponse(sh, *shortenedURLs)
	rawBytes, err := response.MarshalJSON()
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	if contextHasError(w, ctx) {
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", rawBytes)
}

func contextHasError(w http.ResponseWriter, ctx context.Context) bool {
	switch ctx.Err() {
	case context.Canceled:
		fmt.Printf("Request canceled")
		http.Error(w, "Request canceled", http.StatusInternalServerError)
		return true
	case context.DeadlineExceeded:
		fmt.Printf("Request timeout")
		http.Error(w, "Timeout exceeded", http.StatusInternalServerError)
		return true
	}
	return false
}

func mapShortenedURLToExternalResponse(sh *ShortenerHandlers, slice []model.ShortenedURL) ExternalShortenedURLResponseDtoSlice {
	var responseSlice []ExternalShortenedURLResponseDto
	for _, item := range slice {
		responseItem := ExternalShortenedURLResponseDto{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", sh.shortenedURLAddr, item.ShortURL),
		}
		responseSlice = append(responseSlice, responseItem)
	}
	return responseSlice
}

func mapExternalRequestToShortenedURL(slice []ExternalShortenedURLRequestDto) *[]model.ShortenedURL {
	var shortenedURLs []model.ShortenedURL
	for _, item := range slice {
		shortenedURL := model.ShortenedURL{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
		}
		shortenedURLs = append(shortenedURLs, shortenedURL)
	}
	return &shortenedURLs
}
