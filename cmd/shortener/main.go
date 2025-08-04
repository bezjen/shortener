package main

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/router"
	"github.com/bezjen/shortener/internal/service"
	"log"
	"net/http"
)

func main() {
	config.ParseConfig()
	cfg := config.AppConfig

	err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Printf("Error during logger initialization: %v", err)
		return
	}

	storage, err := repository.NewFileRepository(cfg)
	if err != nil {
		log.Printf("Error during storage initialization: %v", err)
		return
	}
	urlShortener := service.NewURLShortener(storage)
	shortenerHandler := handler.NewShortenerHandler(cfg, urlShortener)
	shortenerRouter := router.NewRouter(*shortenerHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := http.ListenAndServe(cfg.ServerAddr, shortenerRouter); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
}
