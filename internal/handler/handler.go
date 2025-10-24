// Package handler provides HTTP handlers for the URL shortening service.
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
	"time"
)

// ShortenerHandler handles HTTP requests for URL shortening operations.
type ShortenerHandler struct {
	cfg          config.Config
	logger       *logger.Logger
	shortener    service.Shortener
	auditService service.AuditService
}

// NewShortenerHandler creates a new instance of ShortenerHandler.
//
// Parameters:
//   - cfg: application configuration settings
//   - logger: logger instance for application logging
//   - shortener: URL shortening service implementation
//   - auditService: service for auditing user actions
//
// Returns:
//   - *ShortenerHandler: initialized HTTP handler
func NewShortenerHandler(
	cfg config.Config,
	logger *logger.Logger,
	shortener service.Shortener,
	auditService service.AuditService,
) *ShortenerHandler {
	return &ShortenerHandler{
		cfg:          cfg,
		logger:       logger,
		shortener:    shortener,
		auditService: auditService,
	}
}

// HandlePostShortURLTextPlain handles POST requests to create short URLs from plain text.
// Accepts the original URL in the request body as text/plain.
//
// Responses:
//   - 201 Created: Short URL successfully created
//   - 409 Conflict: URL was already shortened previously
//   - 400 Bad Request: Invalid URL format
//   - 500 Internal Server Error: Internal server error
//
// Example request:
//
//	POST / HTTP/1.1
//	Content-Type: text/plain
//
//	https://example.com/very-long-url
//
// Example response:
//
//	HTTP/1.1 201 Created
//	Content-Type: text/plain
//
//	http://localhost:8080/abc123def
func (h *ShortenerHandler) HandlePostShortURLTextPlain(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	defer r.Body.Close()

	bodyString, err := h.readBody(r)
	if err != nil {
		h.logger.Error("Failed to read body", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = h.validateURL(bodyString); err != nil {
		http.Error(rw, "incorrect url", http.StatusBadRequest)
		return
	}

	shortURL, err := h.shortener.GenerateShortURLPart(r.Context(), userID, bodyString)
	if err != nil {
		h.handleGenerationError(rw, err, bodyString)
		return
	}

	h.auditEvent(model.ActionShorten, userID, bodyString)
	h.writeTextResponse(rw, http.StatusCreated, shortURL)
}

// HandleGetShortURLRedirect handles GET requests to redirect to original URLs.
// Looks up the original URL by short identifier and performs redirect.
//
// Path parameters:
//   - shortURL: Short URL identifier in the URL path
//
// Responses:
//   - 307 Temporary Redirect: Successful redirect to original URL
//   - 410 Gone: Short URL has been deleted
//   - 400 Bad Request: Missing or invalid short URL parameter
//   - 500 Internal Server Error: Internal server error
//
// Example request:
//
//	GET /abc123def HTTP/1.1
//
// Example response:
//
//	HTTP/1.1 307 Temporary Redirect
//	Location: https://example.com/very-long-url
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

	if resultURL.IsDeleted {
		rw.WriteHeader(http.StatusGone)
		return
	}

	h.auditEvent(model.ActionFollow, getUserIDFromContext(r), resultURL.OriginalURL)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Location", resultURL.OriginalURL)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}

// HandlePostShortURLJSON handles POST requests to create short URLs from JSON.
// Accepts the original URL in a JSON object in the request body.
//
// Request format:
//
//	{"url": "https://example.com/very-long-url"}
//
// Responses:
//   - 201 Created: Short URL successfully created
//   - 409 Conflict: URL was already shortened previously
//   - 400 Bad Request: Invalid JSON or URL format
//   - 500 Internal Server Error: Internal server error
//
// Example response:
//
//	HTTP/1.1 201 Created
//	Content-Type: application/json
//
//	{"result": "http://localhost:8080/abc123def"}
func (h *ShortenerHandler) HandlePostShortURLJSON(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	rw.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var request model.ShortenJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect json")
		return
	}
	if err := h.validateURL(request.URL); err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect url")
		return
	}
	shortURL, err := h.shortener.GenerateShortURLPart(r.Context(), userID, request.URL)
	if err != nil {
		h.handleJSONGenerationError(rw, err, request.URL)
		return
	}

	h.auditEvent(model.ActionShorten, userID, request.URL)
	h.writeShortenJSONSuccessResponse(rw, http.StatusCreated, shortURL)
}

// HandlePostShortURLBatchJSON handles POST requests for batch URL shortening.
// Accepts multiple URLs with correlation IDs and returns shortened versions.
//
// Request format:
//
//	[
//	  {"correlation_id": "1", "original_url": "https://example.com/url1"},
//	  {"correlation_id": "2", "original_url": "https://example.com/url2"}
//	]
//
// Responses:
//   - 201 Created: Batch processing completed successfully
//   - 400 Bad Request: Invalid JSON or URL format
//   - 500 Internal Server Error: Internal server error
//
// Example response:
//
//	HTTP/1.1 201 Created
//	Content-Type: application/json
//
//	[
//	  {"correlation_id": "1", "short_url": "http://localhost:8080/abc123"},
//	  {"correlation_id": "2", "short_url": "http://localhost:8080/def456"}
//	]
func (h *ShortenerHandler) HandlePostShortURLBatchJSON(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	rw.Header().Set("Content-Type", "application/json")

	request, err := h.decodeBatchRequest(r)
	if err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect json")
		return
	}

	for _, requestItem := range request {
		if err = h.validateURL(requestItem.OriginalURL); err != nil {
			h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect url "+requestItem.OriginalURL)
			return
		}
	}

	shortURLs, err := h.shortener.GenerateShortURLPartBatch(r.Context(), userID, request)
	if err != nil {
		h.logger.Error("Failed to generate short URLs batch",
			zap.Error(err),
		)
		h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	response := h.buildBatchResponse(shortURLs)
	h.writeJSONResponse(rw, http.StatusCreated, response)
}

// HandleGetUserURLsJSON handles GET requests to retrieve user's URLs.
// Returns all URLs created by the authenticated user.
//
// Responses:
//   - 200 OK: User URLs retrieved successfully
//   - 204 No Content: User has no shortened URLs
//   - 401 Unauthorized: User not authenticated
//   - 500 Internal Server Error: Internal server error
//
// Example response:
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//
//	[
//	  {"short_url": "http://localhost:8080/abc123", "original_url": "https://example.com/url1"},
//	  {"short_url": "http://localhost:8080/def456", "original_url": "https://example.com/url2"}
//	]
func (h *ShortenerHandler) HandleGetUserURLsJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	userID := getUserIDFromContext(r)
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

	response := h.buildUserURLsResponse(userURLs)
	h.writeJSONResponse(rw, http.StatusOK, response)
}

// HandleDeleteShortURLsBatchJSON handles DELETE requests to mark URLs as deleted.
// Accepts a list of short URLs to mark as deleted (async processing).
//
// Request format:
//
//	["abc123def", "xyz456ghi"]
//
// Responses:
//   - 202 Accepted: Deletion request accepted for processing
//   - 400 Bad Request: Invalid JSON format
//   - 429 Too Many Requests: Deletion queue is full
//
// Example request:
//
//	DELETE /api/user/urls HTTP/1.1
//	Content-Type: application/json
//
//	["abc123def", "xyz456ghi"]
func (h *ShortenerHandler) HandleDeleteShortURLsBatchJSON(rw http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	rw.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		h.writeShortenJSONErrorResponse(rw, http.StatusBadRequest, "incorrect json")
		return
	}

	err := h.shortener.DeleteUserShortURLsBatch(r.Context(), userID, shortURLs)
	if err != nil {
		h.logger.Error("Failed to delete short URLs for user",
			zap.Error(err),
			zap.String("userID", userID),
			zap.Strings("shortURLs", shortURLs),
		)
		h.writeShortenJSONErrorResponse(rw, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		return
	}

	rw.WriteHeader(http.StatusAccepted)
}

// HandlePingRepository handles health check requests to verify storage connectivity.
//
// Responses:
//   - 200 OK: Storage is accessible
//   - 500 Internal Server Error: Storage is unreachable
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

func (h *ShortenerHandler) readBody(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (h *ShortenerHandler) validateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	return err
}

func (h *ShortenerHandler) buildFullURL(shortURL string) (string, error) {
	return url.JoinPath(h.cfg.BaseURL, shortURL)
}

func (h *ShortenerHandler) handleGenerationError(rw http.ResponseWriter, err error, originalURL string) {
	var uniqueURLErr *repository.ErrURLConflict
	if errors.As(err, &uniqueURLErr) {
		resultURL, err := h.buildFullURL(uniqueURLErr.ShortURL)
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

	h.logger.Error("Failed to generate short URL",
		zap.Error(err),
		zap.String("originalURL", originalURL),
	)
	http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *ShortenerHandler) handleJSONGenerationError(rw http.ResponseWriter, err error, originalURL string) {
	var uniqueURLErr *repository.ErrURLConflict
	if errors.As(err, &uniqueURLErr) {
		fullShortURL, err := h.buildFullURL(uniqueURLErr.ShortURL)
		if err != nil {
			h.logger.Error("Failed to generate full short URL",
				zap.Error(err),
				zap.String("baseURL", h.cfg.BaseURL),
				zap.String("shortURL", uniqueURLErr.ShortURL),
			)
			h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		h.writeShortenJSONSuccessResponse(rw, http.StatusConflict, fullShortURL)
		return
	}

	h.logger.Error("Failed to generate short URL",
		zap.Error(err),
		zap.String("originalURL", originalURL),
	)
	h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func (h *ShortenerHandler) writeTextResponse(rw http.ResponseWriter, statusCode int, shortURL string) {
	resultURL, err := h.buildFullURL(shortURL)
	if err != nil {
		h.logger.Error("Failed to build result url", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(statusCode)
	rw.Write([]byte(resultURL))
}

func (h *ShortenerHandler) decodeBatchRequest(r *http.Request) ([]model.ShortenBatchRequestItem, error) {
	var request []model.ShortenBatchRequestItem
	err := json.NewDecoder(r.Body).Decode(&request)
	return request, err
}

func (h *ShortenerHandler) buildBatchResponse(shortURLs []model.ShortenBatchResponseItem) []model.ShortenBatchResponseItem {
	var response []model.ShortenBatchResponseItem
	for _, shortURL := range shortURLs {
		fullShortURL, err := h.buildFullURL(shortURL.ShortURL)
		if err != nil {
			h.logger.Error("Failed to generate full short URL",
				zap.Error(err),
				zap.String("baseURL", h.cfg.BaseURL),
				zap.String("shortURL", shortURL.ShortURL),
				zap.String("correlationID", shortURL.CorrelationID),
			)
			continue
		}
		response = append(response, *model.NewShortenBatchResponseItem(shortURL.CorrelationID, fullShortURL))
	}
	return response
}

func (h *ShortenerHandler) buildUserURLsResponse(userURLs []model.URL) []model.UserURLResponseItem {
	var response []model.UserURLResponseItem
	for _, userURL := range userURLs {
		fullShortURL, err := h.buildFullURL(userURL.ShortURL)
		if err != nil {
			h.logger.Error("Failed to generate full short URL",
				zap.Error(err),
				zap.String("baseURL", h.cfg.BaseURL),
				zap.String("shortURL", userURL.ShortURL),
			)
			continue
		}
		response = append(response, *model.NewUserURLResponseItem(fullShortURL, userURL.OriginalURL))
	}
	return response
}

func (h *ShortenerHandler) writeShortenJSONSuccessResponse(rw http.ResponseWriter, statusCode int, shortURL string) {
	fullURL, err := h.buildFullURL(shortURL)
	if err != nil {
		h.logger.Error("Failed to build full URL", zap.Error(err))
		h.writeShortenJSONErrorResponse(rw, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	h.writeJSONResponse(rw, statusCode, model.ShortenJSONResponse{ShortURL: fullURL})
}

func (h *ShortenerHandler) writeShortenJSONErrorResponse(rw http.ResponseWriter, statusCode int, error string) {
	h.writeJSONResponse(rw, statusCode, model.ShortenJSONResponse{Error: error})
}

func (h *ShortenerHandler) writeJSONResponse(rw http.ResponseWriter, statusCode int, response interface{}) {
	rw.WriteHeader(statusCode)
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func getUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return ""
	}
	return userID
}

func (h *ShortenerHandler) auditEvent(action model.AuditAction, userID string, url string) {
	if h.auditService == nil {
		return
	}

	event := model.NewAuditEvent(time.Now().Unix(), action, userID, url)
	h.auditService.NotifyAll(*event)
}
