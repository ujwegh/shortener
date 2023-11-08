package main

import (
	"fmt"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/router"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"net/http"
	"strings"
)

func main() {
	c := config.ParseFlags()
	logger.InitLogger(c.LogLevel)
	db := storage.Open(c.DatabaseDSN)

	fs := storage.NewFileStorage(c.FileStoragePath)
	dbs := storage.NewDBStorage(db)
	ss := service.NewShortenerService(fs)
	us := handlers.NewShortenerHandlers(c.ShortenedURLAddr, ss, dbs)
	r := router.NewAppRouter(us)

	fmt.Printf("Starting server on port %s...\n", strings.Split(c.ServerAddr, ":")[1])
	http.ListenAndServe(c.ServerAddr, r)
}
