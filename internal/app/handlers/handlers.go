package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type ShortenerHandlers struct {
	urlMap           map[string]string
	shortenedURLAddr string
}

func NewShortenerHandlers(shortenedURLAddr string) *ShortenerHandlers {
	return &ShortenerHandlers{
		urlMap:           make(map[string]string),
		shortenedURLAddr: shortenedURLAddr,
	}
}

func (us *ShortenerHandlers) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	url := string(body)
	if url == "" {
		http.Error(w, "Url is empty", http.StatusBadRequest)
		return
	}
	shortenedURL := generateKey()
	us.urlMap[shortenedURL] = url
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", us.shortenedURLAddr, shortenedURL)
}

func (us *ShortenerHandlers) HandleShortenedURL(w http.ResponseWriter, r *http.Request) {
	shortKey := chi.URLParam(r, "id")
	url, found := us.urlMap[shortKey]
	if !found {
		http.Error(w, "Shortened url not found", http.StatusNotFound)
		return
	}
	w.Header().Add("Location", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
