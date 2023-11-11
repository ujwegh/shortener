package main

import (
	"fmt"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/router"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"github.com/ujwegh/shortener/internal/app/storage/migrations"
	"net/http"
	"strings"
)

func main() {
	c := config.ParseFlags()
	logger.InitLogger(c.LogLevel)

	var us *handlers.ShortenerHandlers
	if c.DatabaseDSN != "" {
		logger.Log.Info("Using database storage.")
		db := storage.Open(c.DatabaseDSN)
		// Migrate the database
		err := storage.MigrateFS(db, migrations.FS, ".")
		if err != nil {
			panic(err)
		}
		dbs := storage.NewDBStorage(db)
		ss := service.NewShortenerService(dbs)
		us = handlers.NewShortenerHandlers(c.ShortenedURLAddr, ss, dbs)
	} else {
		logger.Log.Info("Using in-memory storage.")
		fs := storage.NewFileStorage(c.FileStoragePath)
		ss := service.NewShortenerService(fs)
		us = handlers.NewShortenerHandlers(c.ShortenedURLAddr, ss, fs)
	}

	r := router.NewAppRouter(us)

	fmt.Printf("Starting server on port %s...\n", strings.Split(c.ServerAddr, ":")[1])
	http.ListenAndServe(c.ServerAddr, r)
}
