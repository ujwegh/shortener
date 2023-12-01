package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/middlware"
	"github.com/ujwegh/shortener/internal/app/router"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	c := config.ParseFlags()
	logger.InitLogger(c.LogLevel)
	s := storage.NewStorage(c)
	taskChannel := make(chan service.Task, 100)

	ss := service.NewShortenerService(s, taskChannel)
	sh := handlers.NewShortenerHandlers(c.ShortenedURLAddr, c.ContextTimeoutSec, ss, s)
	ts := service.NewTokenService(c)
	am := middlware.NewAuthMiddleware(ts)

	r := router.NewAppRouter(sh, am)

	// Start the goroutine
	go ss.BatchProcess(serverCtx, taskChannel)
	// The HTTP Server
	server := &http.Server{Addr: c.ServerAddr, Handler: r}

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancelFunc := context.WithTimeout(serverCtx, 30*time.Second)
		cancelFunc()
		// Inform goroutine to stop processing
		close(taskChannel)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	fmt.Printf("Starting server on port %s...\n", strings.Split(c.ServerAddr, ":")[1])
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	// Wait for server context to be stopped
	<-serverCtx.Done()
}
