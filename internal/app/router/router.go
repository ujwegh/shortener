package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/middlware"
)

func NewAppRouter(sh *handlers.ShortenerHandlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlware.RequestLogger)
	r.Use(middlware.ResponseLogger)
	r.Use(middlware.RequestZipper)
	r.Use(middlware.ResponseZipper)

	r.Post("/", sh.ShortenURL)
	r.Get("/ping", sh.Ping)
	r.Post("/api/shorten", sh.APIShortenURL)
	r.Post("/api/shorten/batch", sh.APIShortenURLBatch)
	r.Get("/{id}", sh.HandleShortenedURL)
	return r
}
