package router

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/bezjen/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

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
