// Package router provides HTTP router configuration for the URL shortening service.
// It sets up routes, middleware, and request handling for all API endpoints.
package router

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/bezjen/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the HTTP router with all routes and middleware.
// Sets up authentication, logging, compression, and routes for URL shortening operations.
//
// Parameters:
//   - logger: logger instance for request logging
//   - authorizer: JWT authorizer service for authentication
//   - shortenerHandler: handler for URL shortening operations
//
// Returns:
//   - *chi.Mux: configured HTTP router
//
// Middleware order:
//  1. Authentication - validates JWT tokens and sets user context
//  2. Logging - logs request details and response metrics
//  3. GZIP decompression - decompresses request bodies
//  4. GZIP compression - compresses responses when supported
//
// Routes:
//   - POST / - Create short URL from plain text
//   - GET /ping - Health check endpoint
//   - GET /{shortURL} - Redirect to original URL
//   - POST /api/shorten - Create short URL from JSON
//   - POST /api/shorten/batch - Batch URL shortening
//   - GET /api/user/urls - Get user's URLs
//   - DELETE /api/user/urls - Delete user's URLs
//   - /debug - Profiler endpoint (for development)
func NewRouter(logger *logger.Logger,
	authorizer service.Authorizer,
	shortenerHandler handler.ShortenerHandler,
) *chi.Mux {
	r := chi.NewRouter()
	authMiddleware := middleware.NewAuthMiddleware(authorizer, logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	gzipMiddleware := middleware.NewGzipMiddleware(logger)

	r.Use(
		authMiddleware.WithAuth,
		loggingMiddleware.WithLogging,
		gzipMiddleware.WithGzipRequestDecompression,
		gzipMiddleware.WithGzipResponseCompression)

	r.Post("/", shortenerHandler.HandlePostShortURLTextPlain)
	r.Get("/ping", shortenerHandler.HandlePingRepository)
	r.Get("/{shortURL}", shortenerHandler.HandleGetShortURLRedirect)
	r.Post("/api/shorten", shortenerHandler.HandlePostShortURLJSON)
	r.Post("/api/shorten/batch", shortenerHandler.HandlePostShortURLBatchJSON)
	r.Get("/api/user/urls", shortenerHandler.HandleGetUserURLsJSON)
	r.Delete("/api/user/urls", shortenerHandler.HandleDeleteShortURLsBatchJSON)

	r.Mount("/debug", chimiddleware.Profiler())

	return r
}
