package main

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	"net/http"
)

func main() {
	storage := repository.NewInMemoryRepository()
	urlShortener := service.NewUrlShortener(storage)
	shortenerHandler := handler.NewShortenerHandler(*urlShortener)
	http.HandleFunc(`/`, shortenerHandler.PostShortUrlHandler())
	http.HandleFunc(`/{shortUrl}`, shortenerHandler.GetShortUrlHandler())

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
