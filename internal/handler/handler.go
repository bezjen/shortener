package handler

import (
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	"io"
	"net/http"
)

func PostShortUrlHandler(storage repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			http.Error(rw, "incorrect content type", http.StatusBadRequest)
		}
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, "failed to generate short url", http.StatusInternalServerError)
			return
		}
		bodyString := string(body)
		shortUrl, err := service.GenerateShort(storage)
		if err != nil {
			http.Error(rw, "failed to generate short url", http.StatusInternalServerError)
		}
		storage.Save(*shortUrl, bodyString)

		resultUrl := "http://localhost:8080/" + *shortUrl
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(resultUrl))
	}
}

func GetShortUrlHandler(storage repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			http.Error(rw, "incorrect content type", http.StatusBadRequest)
		}

		shortUrl := ""
		resultUrl := storage.GetByShortUrl(shortUrl)
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusTemporaryRedirect)
		rw.Write([]byte(resultUrl))
	}
}
