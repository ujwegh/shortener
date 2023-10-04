package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type URLShortener struct {
	urlMap map[string]string
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urlMap: make(map[string]string),
	}
}

func (us *URLShortener) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%s", shortenedURL)
}

func (us *URLShortener) HandleShortenedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	shortKey := ""
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) > 1 {
		shortKey = pathParts[len(pathParts)-1]
	}
	url, found := us.urlMap[shortKey]
	if !found {
		http.Error(w, "Shortened url not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Location", url)
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
