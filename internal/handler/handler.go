package handler

import (
	"encoding/json"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type ShortenerHandler struct {
	cfg       config.Config
	logger    *logger.Logger
	shortener service.Shortener
}

func NewShortenerHandler(cfg config.Config, logger *logger.Logger, shortener service.Shortener) *ShortenerHandler {
	return &ShortenerHandler{
		cfg:       cfg,
		logger:    logger,
		shortener: shortener,
	}
}

func (h *ShortenerHandler) HandlePostShortURLTextPlain(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromContext(r)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read body", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	bodyString := string(body)
	_, err = url.ParseRequestURI(bodyString)
	if err != nil {
		http.Error(rw, "incorrect url", http.StatusBadRequest)
		return
	}
	shortURL, err := h.shortener.GenerateShortURLPart(r.Context(), userID, bodyString)
	if err != nil {
		var uniqueURLErr *repository.ErrURLConflict
		if errors.As(err, &uniqueURLErr) {
			resultURL, err := url.JoinPath(h.cfg.BaseURL, uniqueURLErr.ShortURL)
			if err != nil {
				h.logger.Error("Failed to build result url", zap.Error(err))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte(resultURL))
			return
		}

		h.logger.Error("Failed to generate short OriginalURL",
			zap.Error(err),
			zap.String("bodyString", bodyString),
		)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resultURL, err := url.JoinPath(h.cfg.BaseURL, shortURL)
	if err != nil {
		h.logger.Error("Failed to build result url", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	resultURL, err := h.shortener.GetURLByShortURLPart(r.Context(), shortURL)
	if err != nil {
		h.logger.Error("Failed to get url by short url",
			zap.Error(err),
			zap.String("shortURL", shortURL),
		)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Location", resultURL)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) HandlePostShortURLJSON(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromContext(r)
	rw.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var request model.ShortenJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect json")
		return
	}
	if _, err := url.ParseRequestURI(request.URL); err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect url")
		return
	}
	shortURL, err := h.shortener.GenerateShortURLPart(r.Context(), userID, request.URL)
	if err != nil {
		var uniqueURLErr *repository.ErrURLConflict
		if errors.As(err, &uniqueURLErr) {
			fullShortURL, err := url.JoinPath(h.cfg.BaseURL, uniqueURLErr.ShortURL)
			if err != nil {
				h.logger.Error("Failed to generate full short OriginalURL",
					zap.Error(err),
					zap.String("baseURL", h.cfg.BaseURL),
					zap.String("shortURL", shortURL),
				)
				h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
			h.writeShortenJSONSuccessResponse(rw, http.StatusConflict, fullShortURL)
			return
		}

		h.logger.Error("Failed to generate short OriginalURL",
			zap.Error(err),
			zap.String("originalURL", request.URL),
		)
		h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	fullShortURL, err := url.JoinPath(h.cfg.BaseURL, shortURL)
	if err != nil {
		h.logger.Error("Failed to generate full short OriginalURL",
			zap.Error(err),
			zap.String("baseURL", h.cfg.BaseURL),
			zap.String("shortURL", shortURL),
		)
		h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	h.writeShortenJSONSuccessResponse(rw, http.StatusCreated, fullShortURL)
}

func (h *ShortenerHandler) HandlePostShortURLBatchJSON(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromContext(r)
	rw.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var request []model.ShortenBatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect json")
		return
	}
	for _, requestItem := range request {
		if _, err := url.ParseRequestURI(requestItem.OriginalURL); err != nil {
			h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect url "+requestItem.OriginalURL)
			return
		}
	}
	shortURLs, err := h.shortener.GenerateShortURLPartBatch(r.Context(), userID, request)
	if err != nil {
		h.logger.Error("Failed to generate short OriginalURL",
			zap.Error(err),
		)
		h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	var response []model.ShortenBatchResponseItem
	for _, shortURL := range shortURLs {
		fullShortURL, err := url.JoinPath(h.cfg.BaseURL, shortURL.ShortURL)
		if err != nil {
			h.logger.Error("Failed to generate full short OriginalURL",
				zap.Error(err),
				zap.String("baseURL", h.cfg.BaseURL),
				zap.String("shortURL", shortURL.ShortURL),
				zap.String("correlationID", shortURL.CorrelationID),
			)
			h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		response = append(response, *model.NewShortenBatchResponseItem(shortURL.CorrelationID, fullShortURL))
	}

	h.writeJSONBatchResponse(rw, http.StatusCreated, response)
}

func (h *ShortenerHandler) HandleGetUserURLsJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	userID := getUserIdFromContext(r)
	if userID == "" {
		h.writeShortenJSONErrorResponse(rw, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	defer r.Body.Close()

	userURLs, err := h.shortener.GetURLsByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get urls for user", zap.Error(err), zap.String("userID", userID))
		h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if len(userURLs) == 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	var response []model.UserUrlResponseItem
	for _, userURL := range userURLs {
		fullShortURL, err := url.JoinPath(h.cfg.BaseURL, userURL.ShortURL)
		if err != nil {
			h.logger.Error("Failed to generate full short OriginalURL",
				zap.Error(err),
				zap.String("baseURL", h.cfg.BaseURL),
				zap.String("shortURL", userURL.ShortURL),
			)
			h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		response = append(response, *model.NewUserUrlResponseItem(fullShortURL, userURL.OriginalURL))
	}

	h.writeJSONUserURLsResponse(rw, http.StatusCreated, response)
}

func (h *ShortenerHandler) HandlePingRepository(rw http.ResponseWriter, r *http.Request) {
	err := h.shortener.PingRepository(r.Context())
	if err != nil {
		h.logger.Error("Failed to ping",
			zap.Error(err),
		)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
}

func (h *ShortenerHandler) writeShortenJSONSuccessResponse(rw http.ResponseWriter, statusCode int, shortURL string) {
	h.writeJSONResponse(rw, statusCode, model.ShortenJSONResponse{ShortURL: shortURL})
}

func (h *ShortenerHandler) writeShortenJSONErrorResponse(rw http.ResponseWriter, statusCode int, error string) {
	h.writeJSONResponse(rw, statusCode, model.ShortenJSONResponse{Error: error})
}

func (h *ShortenerHandler) writeJSONResponse(rw http.ResponseWriter, statusCode int, response model.ShortenJSONResponse) {
	rw.WriteHeader(statusCode)
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode error response", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *ShortenerHandler) writeJSONBatchResponse(rw http.ResponseWriter,
	statusCode int, response []model.ShortenBatchResponseItem,
) {
	rw.WriteHeader(statusCode)
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode error response", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *ShortenerHandler) writeJSONUserURLsResponse(rw http.ResponseWriter,
	statusCode int, response []model.UserUrlResponseItem,
) {
	rw.WriteHeader(statusCode)
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode error response", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func getUserIdFromContext(r *http.Request) string {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return ""
	}
	return userID
}
