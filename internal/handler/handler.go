package handler

import (
	"github.com/bezjen/shortener/internal/service"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	if !strings.HasPrefix(contentType, "text/plain") {
		http.Error(rw, "incorrect content type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "failed to read body", http.StatusInternalServerError)
		return
	}
	bodyString := string(body)
	_, errValidation := url.ParseRequestURI(bodyString)
	if errValidation != nil {
		http.Error(rw, "failed to parse url", http.StatusBadRequest)
		return
	}
	shortUrl, err := h.shortener.GenerateShortUrlPart(bodyString)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	resultUrl := "http://localhost:8080/" + shortUrl
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(resultUrl))
}

func (h *ShortenerHandler) handleGetShortUrl(rw http.ResponseWriter, r *http.Request) {
	// TODO: check if content type validation required
	// Request example:
	// GET /EwHXdJfB HTTP/1.1
	// Host: localhost:8080
	// Content-Type: text/plain
	//contentType := r.Header.Get("Content-Type")
	//if !strings.HasPrefix(contentType, "text/plain") {
	//	http.Error(rw, "incorrect content type", http.StatusBadRequest)
	//	return
	//}

	shortUrl := r.URL.Path[1:]
	if shortUrl == "" {
		http.Error(rw, "short url is empty", http.StatusBadRequest)
		return
	}
	resultUrl, err := h.shortener.GetUrlByShortUrlPart(shortUrl)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Location", resultUrl)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}
