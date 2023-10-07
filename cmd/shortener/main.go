package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"net/http"
)

type Server struct {
	Router    *chi.Mux
	shortener handlers.URLShortener
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	s.shortener = *handlers.NewURLShortener()

	return s
}

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.Logger)

	s.Router.Get("/", s.shortener.ShortenURL)
	s.Router.Post("/{id}", s.shortener.HandleShortenedURL)
}

func main() {
	s := CreateNewServer()
	s.MountHandlers()
	fmt.Println("Starting server on port 8080")
	err := http.ListenAndServe(`:8080`, s.Router)
	if err != nil {
		panic(err)
	}
}
