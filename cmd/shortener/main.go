package main

import (
	"fmt"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/router"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func main() {
	c := config.ParseFlags()
	initLogger(c.LogLevel)
	s := storage.NewFileStorage(c.FileStoragePath)
	ss := service.NewShortenerService(s)
	us := handlers.NewShortenerHandlers(c.ShortenedURLAddr, ss)
	r := router.NewAppRouter(us)
	fmt.Printf("Starting server on port %s...\n", strings.Split(c.ServerAddr, ":")[1])
	http.ListenAndServe(c.ServerAddr, r)
}

func initLogger(level string) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		panic(err)
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	logger.Initialize(zl)
}
