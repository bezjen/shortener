package middleware

import (
	"bytes"
	"github.com/bezjen/shortener/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewLoggingMiddleware(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	middleware := NewLoggingMiddleware(testLogger)

	if middleware == nil {
		t.Error("Expected middleware to be created")
	}
}

func TestWithLogging(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	middleware := NewLoggingMiddleware(testLogger)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := middleware.WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestLoggingResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: rr,
		responseData:   responseData,
	}

	data := []byte("test data")
	size, err := lw.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if size != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), size)
	}
	if responseData.size != len(data) {
		t.Errorf("Expected responseData size %d, got %d", len(data), responseData.size)
	}
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: rr,
		responseData:   responseData,
	}

	lw.WriteHeader(http.StatusNotFound)

	if responseData.status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, responseData.status)
	}
}

func TestWithLogging_ResponseData(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	middleware := NewLoggingMiddleware(testLogger)

	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("request body"))
	rr := httptest.NewRecorder()

	handler := middleware.WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("response body"))
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}
}
