package handler

import (
	"github.com/bezjen/shortener/internal/service"
	"io"
	"net/http"
)

type ShortenerHandler struct {
	shortener service.UrlShortener
}

func NewShortenerHandler(shortener service.UrlShortener) *ShortenerHandler {
	return &ShortenerHandler{
		shortener: shortener,
	}
}

func (h *ShortenerHandler) PostShortUrlHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			http.Error(rw, "incorrect content type", http.StatusBadRequest)
		}
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, "failed to read body", http.StatusInternalServerError)
			return
		}
		bodyString := string(body)
		shortUrl, err := h.shortener.GenerateShortUrlPart(bodyString)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		resultUrl := "http://localhost:8080/" + shortUrl
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(resultUrl))
	}
}

func (h *ShortenerHandler) GetShortUrlHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			http.Error(rw, "incorrect content type", http.StatusBadRequest)
		}

		shortUrl := ""
		resultUrl, err := h.shortener.GetUrlByShortUrlPart(shortUrl)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusTemporaryRedirect)
		rw.Write([]byte(resultUrl))
	}
}
