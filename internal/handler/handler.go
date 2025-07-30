package handler

import (
	"encoding/json"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ShortenerHandler struct {
	shortener service.Shortener
	cfg       config.Config
}

func NewShortenerHandler(cfg config.Config, shortener service.Shortener) *ShortenerHandler {
	return &ShortenerHandler{
		cfg:       cfg,
		shortener: shortener,
	}
}

func (h *ShortenerHandler) HandlePostShortURLTextPlain(rw http.ResponseWriter, r *http.Request) {
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
	_, err = url.ParseRequestURI(bodyString)
	if err != nil {
		http.Error(rw, "incorrect url", http.StatusBadRequest)
		return
	}
	shortURL, err := h.shortener.GenerateShortURLPart(bodyString)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	resultURL := h.cfg.BaseURL + "/" + shortURL
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(resultURL))
}

func (h *ShortenerHandler) HandleGetShortURLRedirect(rw http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	if shortURL == "" {
		http.Error(rw, "short url is empty", http.StatusBadRequest)
		return
	}
	resultURL, err := h.shortener.GetURLByShortURLPart(shortURL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Location", resultURL)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) HandlePostShortURLJSON(rw http.ResponseWriter, r *http.Request) {
	var result model.PostShortURLJSONResponse
	contentType := r.Header.Get("Content-Type")
	rw.Header().Set("Content-Type", "application/json")
	if !strings.HasPrefix(contentType, "application/json") {
		result.Error = "incorrect content type"
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(result)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		result.Error = "failed to read body"
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(result)
		return
	}
	var request model.PostShortURLJSONRequest
	if err = json.Unmarshal(body, &request); err != nil {
		result.Error = "incorrect json"
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(result)
		return
	}
	_, err = url.ParseRequestURI(request.URL)
	if err != nil {
		result.Error = "incorrect url"
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(result)
		return
	}

	shortURL, err := h.shortener.GenerateShortURLPart(request.URL)
	if err != nil {
		result.Error = err.Error()
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(result)
		return
	}

	result.ShortURL = h.cfg.BaseURL + "/" + shortURL
	rw.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
