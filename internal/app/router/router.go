package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/middlware"
)

func NewAppRouter(us *handlers.ShortenerHandlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlware.RequestLogger)
	r.Use(middlware.ResponseLogger)
	r.Use(middlware.RequestZipper)
	r.Use(middlware.ResponseZipper)

	r.Post("/", us.ShortenURL)
	r.Get("/ping", us.Ping)
	r.Post("/api/shorten", us.APIShortenURL)
	r.Get("/{id}", us.HandleShortenedURL)
	return r
}
