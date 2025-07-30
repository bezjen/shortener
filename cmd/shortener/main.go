package main

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	config.ParseConfig()
	cfg := config.AppConfig
	storage := repository.NewInMemoryRepository()
	urlShortener := service.NewURLShortener(storage)
	shortenerHandler := handler.NewShortenerHandler(cfg, urlShortener)

	r := chi.NewRouter()
	r.Post("/", shortenerHandler.HandlePostShortURL)
	r.Get("/{shortURL}", shortenerHandler.HandleGetShortURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := http.ListenAndServe(cfg.ServerAddr, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
}
