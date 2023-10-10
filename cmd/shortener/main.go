package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"net/http"
	"strings"
)

func main() {
	appConfig := config.ParseFlags()
	if err := run(appConfig); err != nil {
		panic(err)
	}
}

func run(config config.AppConfig) error {
	r := chi.NewRouter()
	us := handlers.NewURLShortener(config.ShortenedURLAddr)

	r.Use(middleware.Logger)

	r.Post("/", us.ShortenURL)
	r.Get("/{id}", us.HandleShortenedURL)

	fmt.Printf("Starting server on port %s...\n", strings.Split(config.ServerAddr, ":")[1])
	return http.ListenAndServe(config.ServerAddr, r)
}
