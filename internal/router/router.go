package router

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(logger *logger.Logger, shortenerHandler handler.ShortenerHandler) *chi.Mux {
	r := chi.NewRouter()
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	gzipMiddleware := middleware.NewGzipMiddleware(logger)

	r.Use(
		loggingMiddleware.WithLogging,
		gzipMiddleware.WithGzipRequestDecompression,
		gzipMiddleware.WithGzipResponseCompression)

	r.Post("/", shortenerHandler.HandlePostShortURLTextPlain)
	r.Get("/ping", shortenerHandler.HandlePingRepository)
	r.Get("/{shortURL}", shortenerHandler.HandleGetShortURLRedirect)
	r.Post("/api/shorten", shortenerHandler.HandlePostShortURLJSON)
	r.Post("/api/shorten/batch", shortenerHandler.HandlePostShortURLBatchJSON)

	return r
}
