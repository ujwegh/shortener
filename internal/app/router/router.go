package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/middlware"
	"github.com/ujwegh/shortener/internal/app/storage"
)

func NewAppRouter(config config.AppConfig) *chi.Mux {
	r := chi.NewRouter()
	s := storage.NewFileStorage(config.FileStoragePath)
	us := handlers.NewShortenerHandlers(config.ShortenedURLAddr, s)
	r.Use(logger.RequestLogger)
	r.Use(logger.ResponseLogger)
	r.Use(middlware.RequestZipper)
	r.Use(middlware.ResponseZipper)

	r.Post("/", us.ShortenURL)
	r.Post("/api/shorten", us.APIShortenURL)
	r.Get("/{id}", us.HandleShortenedURL)
	return r
}
