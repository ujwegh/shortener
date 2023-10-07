package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	us := handlers.NewURLShortener()

	r.Use(middleware.Logger)

	r.Post("/", us.ShortenURL)
	r.Get("/{id}", us.HandleShortenedURL)

	fmt.Println("Starting server on port 8080")
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
