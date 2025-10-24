// Package middleware provides HTTP middleware components for the URL shortening service.
package middleware

import (
	"github.com/bezjen/shortener/internal/compress"
	"github.com/bezjen/shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// GzipMiddleware provides GZIP compression and decompression for HTTP requests.
// It automatically compresses responses and decompresses requests when appropriate.
type GzipMiddleware struct {
	logger *logger.Logger
}

// NewGzipMiddleware creates a new GzipMiddleware instance.
//
// Parameters:
//   - logger: logger instance for logging compression events
//
// Returns:
//   - *GzipMiddleware: initialized GZIP middleware
func NewGzipMiddleware(logger *logger.Logger) *GzipMiddleware {
	return &GzipMiddleware{logger: logger}
}

// WithGzipRequestDecompression wraps an HTTP handler with GZIP request decompression.
// Automatically decompresses request bodies with Content-Encoding: gzip header.
//
// Parameters:
//   - h: HTTP handler to wrap
//
// Returns:
//   - http.Handler: wrapped handler that decompresses GZIP requests
func (m *GzipMiddleware) WithGzipRequestDecompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gr, err := compress.NewGzipReader(r.Body)
		if err != nil {
			m.logger.Error("Failed to decompress request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gr *compress.GzipReader) {
			err = gr.Close()
			if err != nil {
				m.logger.Error("Failed to close gzip reader", zap.Error(err))
			}
		}(gr)
		r.Body = gr
		h.ServeHTTP(w, r)
	})
}

// WithGzipResponseCompression wraps an HTTP handler with GZIP response compression.
// Automatically compresses responses for clients that Accept-Encoding: gzip.
// Skips compression for DELETE requests and when client doesn't support GZIP.
//
// Parameters:
//   - h: HTTP handler to wrap
//
// Returns:
//   - http.Handler: wrapped handler that compresses responses
func (m *GzipMiddleware) WithGzipResponseCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			h.ServeHTTP(w, r)
			return
		}

		gw := compress.NewGzipWriter(w)
		defer func(gw *compress.GzipWriter) {
			err := gw.Close()
			if err != nil {
				m.logger.Error("Failed to close gzip writer", zap.Error(err))
			}
		}(gw)
		h.ServeHTTP(gw, r)
	})
}
