package main

import (
	"fmt"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/router"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func main() {
	c := config.ParseFlags()
	initLogger(c.LogLevel)
	r := router.NewAppRouter(c)
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
