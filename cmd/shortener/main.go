package main

import (
	"fmt"
	app "github.com/ujwegh/shortener/internal/app"
	"net/http"
)

func main() {
	shortener := app.NewURLShortener()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			shortener.ShortenURL(w, r)
			return
		}
		shortener.HandleShortenedURL(w, r)
	})

	fmt.Println("Starting server on port 8080")
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}