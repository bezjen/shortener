package router

import (
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(shortenerHandler handler.ShortenerHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.WithLogging,
		middleware.WithGzipRequestDecompression,
		middleware.WithGzipResponseCompression)

	r.Post("/", shortenerHandler.HandlePostShortURLTextPlain)
	r.Get("/{shortURL}", shortenerHandler.HandleGetShortURLRedirect)
	r.Post("/api/shorten", shortenerHandler.HandlePostShortURLJSON)

	return r
}
