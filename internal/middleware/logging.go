// Package middleware provides HTTP middleware components for the URL shortening service.
// Includes logging, authentication, and other request processing middleware.
package middleware

import (
	"github.com/bezjen/shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// LoggingMiddleware provides HTTP request logging functionality.
// It logs request details including URI, method, duration, status code, and response size.
type LoggingMiddleware struct {
	logger *logger.Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware instance.
//
// Parameters:
//   - logger: logger instance for writing log entries
//
// Returns:
//   - *LoggingMiddleware: initialized logging middleware
func NewLoggingMiddleware(logger *logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// WithLogging wraps an HTTP handler with request logging.
// Logs request details before and after handling the request.
//
// Parameters:
//   - h: HTTP handler to wrap with logging
//
// Returns:
//   - http.Handler: wrapped handler that logs request details
func (m *LoggingMiddleware) WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		response := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   response,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		m.logger.Infoln("got incoming HTTP request",
			zap.String("URI", uri),
			zap.String("method", method),
			zap.String("duration", duration.String()),
			zap.Int("status", response.status),
			zap.Int("size", response.size),
		)
	})
}

// responseData holds HTTP response metadata for logging.
// Tracks status code and response size.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter wraps http.ResponseWriter to capture response data.
// Intercepts Write and WriteHeader calls to collect metrics for logging.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write intercepts response writing to capture the response size.
// Implements the http.ResponseWriter interface.
//
// Parameters:
//   - b: byte slice to write
//
// Returns:
//   - int: number of bytes written
//   - error: write error if any
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader intercepts status code setting to capture the response status.
// Implements the http.ResponseWriter interface.
//
// Parameters:
//   - statusCode: HTTP status code to set
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
