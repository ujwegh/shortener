package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
)

func NewAppRouter(config config.AppConfig) *chi.Mux {
	r := chi.NewRouter()
	us := handlers.NewShortenerHandlers(config.ShortenedURLAddr)
	//r.Use(middleware.Logger)
	r.Use(logger.RequestLogger)
	r.Use(logger.ResponseLogger)

	r.Post("/", us.ShortenURL)
	r.Get("/{id}", us.HandleShortenedURL)
	return r
}
