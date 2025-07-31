package middleware

import (
	"github.com/bezjen/shortener/internal/compress"
	"github.com/bezjen/shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		logger.Log.Infoln("got incoming HTTP request",
			zap.String("URI", uri),
			zap.String("method", method),
			zap.String("duration", duration.String()),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
}

func WithGzipCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// TODO: decompress
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gw := compress.NewGzipWriter(w)
			defer gw.Close()
			h.ServeHTTP(gw, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
