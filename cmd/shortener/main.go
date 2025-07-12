package main

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/repository"
	"net/http"
)

func main() {
	storage := repository.NewInMemoryRepository()
	http.HandleFunc(`/`, handler.PostShortUrlHandler(storage))
	http.HandleFunc(`/{shortUrl}`, handler.GetShortUrlHandler(storage))

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
