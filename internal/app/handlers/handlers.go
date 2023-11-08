package handlers

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
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
	}
	//easyjson:json
	ShortenRequestDto struct {
		URL string `json:"url"`
	}
	//easyjson:json
	ShortenResponseDto struct {
		Result string `json:"result"`
	}
)

func NewShortenerHandlers(shortenedURLAddr string, service service.ShortenerService, storage storage.Storage) *ShortenerHandlers {
	return &ShortenerHandlers{
		shortenerService: service,
		storage:          storage,
		shortenedURLAddr: shortenedURLAddr,
	}
}

func (us *ShortenerHandlers) ShortenURL(w http.ResponseWriter, r *http.Request) {
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
	shortenedURL, err := us.shortenerService.CreateShortenedURL(originalURL)
	if err != nil {
		http.Error(w, "Unable to create shortened URL", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", us.shortenedURLAddr, shortenedURL.ShortURL)
}

func (us *ShortenerHandlers) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	request := ShortenRequestDto{}
	err = easyjson.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	if request.URL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}
	shortenedURL, err := us.shortenerService.CreateShortenedURL(request.URL)
	if err != nil {
		http.Error(w, "Unable to create shortened URL", http.StatusInternalServerError)
		return
	}
	response := &ShortenResponseDto{Result: fmt.Sprintf("%s/%s", us.shortenedURLAddr, shortenedURL.ShortURL)}
	rawBytes, err := easyjson.Marshal(response)
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", rawBytes)
}

func (us *ShortenerHandlers) HandleShortenedURL(w http.ResponseWriter, r *http.Request) {
	shortKey := chi.URLParam(r, "id")
	shortenedURL, err := us.shortenerService.GetShortenedURL(shortKey)
	if err != nil {
		http.Error(w, "Unable to get shortened URL", http.StatusInternalServerError)
		return
	}
	originalURL := shortenedURL.OriginalURL
	if originalURL == "" {
		http.Error(w, "Shortened url not found", http.StatusNotFound)
		return
	}
	w.Header().Add("Location", originalURL)
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (us *ShortenerHandlers) Ping(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := us.storage.Ping(ctx)
	if err != nil {
		http.Error(writer, "Unable to ping storage", http.StatusInternalServerError)
		return
	}

	switch ctx.Err() {
	case context.Canceled:
		http.Error(writer, "Request canceled", http.StatusInternalServerError)
		return
	case context.DeadlineExceeded:
		http.Error(writer, "Timeout exceeded", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
