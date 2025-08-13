package router

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(logger *logger.Logger, shortenerHandler handler.ShortenerHandler) *chi.Mux {
	r := chi.NewRouter()
	m := middleware.NewMiddleware(logger)

	r.Use(
		m.WithLogging,
		m.WithGzipRequestDecompression,
		m.WithGzipResponseCompression)

	r.Post("/", shortenerHandler.HandlePostShortURLTextPlain)
	r.Get("/ping", shortenerHandler.HandlePingRepository)
	r.Get("/{shortURL}", shortenerHandler.HandleGetShortURLRedirect)
	r.Post("/api/shorten", shortenerHandler.HandlePostShortURLJSON)

	return r
}
