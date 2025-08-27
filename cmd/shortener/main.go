package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/config/db"
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

	shortenerLogger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Printf("Error during logger initialization: %v", err)
		return
	}

	storage, err := db.InitDB(cfg)
	if err != nil {
		log.Printf("Error during storage initialization: %v", err)
		return
	}
	defer func(storage repository.Repository) {
		err := storage.Close()
		if err != nil {
			log.Printf("Error during storage close cleanly: %v", err)
		}
	}(storage)
	urlShortener := service.NewURLShortener(storage)
	secretKey, err := generateRandomKey(32)
	if err != nil {
		log.Fatal("Failed to generate secret key:", err)
	}
	authorizer := service.NewAuthorizer(secretKey, shortenerLogger)
	shortenerHandler := handler.NewShortenerHandler(cfg, shortenerLogger, urlShortener)
	shortenerRouter := router.NewRouter(shortenerLogger, authorizer, *shortenerHandler)

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

func generateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}
