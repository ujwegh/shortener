package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/middlware"
	"github.com/ujwegh/shortener/internal/app/storage"
)

func NewAppRouter(config config.AppConfig) *chi.Mux {
	r := chi.NewRouter()
	s := storage.NewFileStorage(config.FileStoragePath)
	us := handlers.NewShortenerHandlers(config.ShortenedURLAddr, s)
	r.Use(middlware.RequestLogger)
	r.Use(middlware.ResponseLogger)
	r.Use(middlware.RequestZipper)
	r.Use(middlware.ResponseZipper)
	r.Use()

	r.Post("/", us.ShortenURL)
	r.Post("/api/shorten", us.APIShortenURL)
	r.Get("/{id}", us.HandleShortenedURL)
	return r
}
