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

func (h *ShortenerHandler) HandleMainPage() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.handlePostShortUrl(rw, r)
		case http.MethodGet:
			h.handleGetShortUrl(rw, r)
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (h *ShortenerHandler) handlePostShortUrl(rw http.ResponseWriter, r *http.Request) {
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

func (h *ShortenerHandler) handleGetShortUrl(rw http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(rw, "incorrect content type", http.StatusBadRequest)
	}

	shortUrl := r.URL.Path[1:]
	if shortUrl == "" {
		http.Error(rw, "short url is empty", http.StatusBadRequest)
		return
	}
	resultUrl, err := h.shortener.GetUrlByShortUrlPart(shortUrl)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusTemporaryRedirect)
	rw.Write([]byte(resultUrl))
}
