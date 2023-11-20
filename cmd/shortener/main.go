package main

import (
	"fmt"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/middlware"
	"github.com/ujwegh/shortener/internal/app/router"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"net/http"
	"strings"
)

func main() {
	c := config.ParseFlags()
	logger.InitLogger(c.LogLevel)
	s := storage.NewStorage(c)
	ss := service.NewShortenerService(s)
	sh := handlers.NewShortenerHandlers(c.ShortenedURLAddr, c.ContextTimeoutSec, ss, s)
	ts := service.NewTokenService(c)
	am := middlware.NewAuthMiddleware(ts)

	r := router.NewAppRouter(sh, am)

	fmt.Printf("Starting server on port %s...\n", strings.Split(c.ServerAddr, ":")[1])
	http.ListenAndServe(c.ServerAddr, r)
}
