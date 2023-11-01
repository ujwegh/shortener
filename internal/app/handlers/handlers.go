package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/storage"
	"io"
	"net/http"
)

type ShortenerHandlers struct {
	storage          storage.Storage
	shortenedURLAddr string
}

func NewShortenerHandlers(shortenedURLAddr string, s storage.Storage) *ShortenerHandlers {
	return &ShortenerHandlers{
		storage:          s,
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
	shortURL := generateKey()
	shortenedURL := &model.ShortenedURL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	err = us.storage.WriteShortenedURL(shortenedURL)
	if err != nil {
		http.Error(w, "Unable to write shortened URL", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", us.shortenedURLAddr, shortURL)
}

func (us *ShortenerHandlers) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	request := model.ShortenRequestDto{}
	err = easyjson.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	if request.URL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}
	shortURL := generateKey()
	shortenedURL := &model.ShortenedURL{
		ShortURL:    shortURL,
		OriginalURL: request.URL,
	}
	err = us.storage.WriteShortenedURL(shortenedURL)
	if err != nil {
		http.Error(w, "Unable to write shortened URL", http.StatusInternalServerError)
		return
	}
	response := &model.ShortenResponseDto{Result: fmt.Sprintf("%s/%s", us.shortenedURLAddr, shortenedURL.ShortURL)}
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
	shortenedURL, err := us.storage.ReadShortenedURL(shortKey)
	if err != nil {
		http.Error(w, "Unable to read shortened URL", http.StatusInternalServerError)
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

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
