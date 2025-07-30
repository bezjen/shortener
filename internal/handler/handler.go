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
	rw.Header().Set("Content-Type", "application/json")

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeJSONErrorResponse(rw, http.StatusBadRequest, "incorrect content type")
		return
	}
	defer r.Body.Close()
	var request model.PostShortURLJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONErrorResponse(rw, http.StatusBadRequest, "incorrect json")
		return
	}
	if _, err := url.ParseRequestURI(request.URL); err != nil {
		writeJSONErrorResponse(rw, http.StatusBadRequest, "incorrect url")
		return
	}
	shortURL, err := h.shortener.GenerateShortURLPart(request.URL)
	if err != nil {
		writeJSONErrorResponse(rw, http.StatusInternalServerError, err.Error())
		return
	}

	fullShortURL := h.cfg.BaseURL + "/" + shortURL
	writeJSONSuccessResponse(rw, http.StatusCreated, fullShortURL)
}

func writeJSONSuccessResponse(rw http.ResponseWriter, statusCode int, shortURL string) {
	writeJSONResponse(rw, statusCode, model.PostShortURLJSONResponse{ShortURL: shortURL})
}

func writeJSONErrorResponse(rw http.ResponseWriter, statusCode int, error string) {
	writeJSONResponse(rw, statusCode, model.PostShortURLJSONResponse{Error: error})
}

func writeJSONResponse(rw http.ResponseWriter, statusCode int, response model.PostShortURLJSONResponse) {
	rw.WriteHeader(statusCode)
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		http.Error(rw, "failed to encode error response", http.StatusInternalServerError)
	}
}
