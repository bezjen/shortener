package main

import (
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	storage := repository.NewInMemoryRepository()
	urlShortener := service.NewURLShortener(storage)
	shortenerHandler := handler.NewShortenerHandler(urlShortener)

	r := chi.NewRouter()
	r.Post("/", shortenerHandler.HandlePostShortURL)
	r.Get("/{shortURL}", shortenerHandler.HandleGetShortURL)

	err := http.ListenAndServe(config.FlagRunAddr, r)
	if err != nil {
		panic(err)
	}
}
