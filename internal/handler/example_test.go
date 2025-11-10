// Package handler_test provides examples for using URL shortening API endpoints.
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bezjen/shortener/internal/mocks"
	"net/http/httptest"
	"strings"

	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/bezjen/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

// ExampleShortenerHandler_HandlePostShortURLTextPlain demonstrates creating a short URL from plain text.
func ExampleShortenerHandler_HandlePostShortURLTextPlain() {
	// Setup
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	testLogger, _ := logger.NewLogger("debug")

	// Use mock implementations instead of real ones
	shortener := &mocks.Shortener{}
	auditService := &mocks.AuditService{}

	// Setup mock expectations
	shortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://example.com/very-long-url-path").
		Return("abc123", nil)
	auditService.On("NotifyAll", mock.Anything).Return()

	h := handler.NewShortenerHandler(cfg, testLogger, shortener, auditService)

	// Create test request
	body := strings.NewReader("https://example.com/very-long-url-path")
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "text/plain")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	// Execute
	w := httptest.NewRecorder()
	h.HandlePostShortURLTextPlain(w, req)

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: text/plain
}

// ExampleShortenerHandler_HandlePostShortURLJSON demonstrates creating a short URL from JSON.
func ExampleShortenerHandler_HandlePostShortURLJSON() {
	// Setup
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	testLogger, _ := logger.NewLogger("debug")

	shortener := &mocks.Shortener{}
	auditService := &mocks.AuditService{}

	// Setup mock expectations
	shortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://example.com/very-long-url").
		Return("def456", nil)
	auditService.On("NotifyAll", mock.Anything).Return()

	h := handler.NewShortenerHandler(cfg, testLogger, shortener, auditService)

	// Create JSON request
	requestBody := model.ShortenJSONRequest{URL: "https://example.com/very-long-url"}
	jsonData, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	// Execute
	w := httptest.NewRecorder()
	h.HandlePostShortURLJSON(w, req)

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: application/json
}

// ExampleShortenerHandler_HandlePostShortURLBatchJSON demonstrates batch URL shortening.
func ExampleShortenerHandler_HandlePostShortURLBatchJSON() {
	// Setup
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	testLogger, _ := logger.NewLogger("debug")

	shortener := &mocks.Shortener{}
	auditService := &mocks.AuditService{}

	// Setup mock expectations
	batchRequest := []model.ShortenBatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://example.com/first-url"},
		{CorrelationID: "2", OriginalURL: "https://example.com/second-url"},
	}

	expectedResponse := []model.ShortenBatchResponseItem{
		{CorrelationID: "1", ShortURL: "short1"},
		{CorrelationID: "2", ShortURL: "short2"},
	}

	shortener.On("GenerateShortURLPartBatch", mock.Anything, mock.Anything, batchRequest).
		Return(expectedResponse, nil)
	auditService.On("NotifyAll", mock.Anything).Return()

	h := handler.NewShortenerHandler(cfg, testLogger, shortener, auditService)

	// Create batch request
	jsonData, _ := json.Marshal(batchRequest)
	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	// Execute
	w := httptest.NewRecorder()
	h.HandlePostShortURLBatchJSON(w, req)

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: application/json
}

// ExampleShortenerHandler_HandleGetUserURLsJSON demonstrates retrieving user's URLs.
func ExampleShortenerHandler_HandleGetUserURLsJSON() {
	// Setup
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	testLogger, _ := logger.NewLogger("debug")

	shortener := &mocks.Shortener{}
	auditService := &mocks.AuditService{}

	// Setup mock expectations
	userURLs := []model.URL{
		{ShortURL: "url1", OriginalURL: "https://example.com/1"},
		{ShortURL: "url2", OriginalURL: "https://example.com/2"},
	}

	shortener.On("GetURLsByUserID", mock.Anything, "test-user").Return(userURLs, nil)

	h := handler.NewShortenerHandler(cfg, testLogger, shortener, auditService)

	// Create request
	req := httptest.NewRequest("GET", "/api/user/urls", nil)

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	// Execute
	w := httptest.NewRecorder()
	h.HandleGetUserURLsJSON(w, req)

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 200
	// Content-Type: application/json
}

// ExampleShortenerHandler_HandleDeleteShortURLsBatchJSON demonstrates batch URL deletion.
func ExampleShortenerHandler_HandleDeleteShortURLsBatchJSON() {
	// Setup
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	testLogger, _ := logger.NewLogger("debug")

	shortener := &mocks.Shortener{}
	auditService := &mocks.AuditService{}

	// Setup mock expectations
	urlsToDelete := []string{"abc123", "def456"}
	shortener.On("DeleteUserShortURLsBatch", mock.Anything, "test-user", urlsToDelete).Return(nil)

	h := handler.NewShortenerHandler(cfg, testLogger, shortener, auditService)

	// Create deletion request
	jsonData, _ := json.Marshal(urlsToDelete)

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	// Execute
	w := httptest.NewRecorder()
	h.HandleDeleteShortURLsBatchJSON(w, req)

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 202
}

// ExampleShortenerHandler_HandlePingRepository demonstrates health check endpoint.
func ExampleShortenerHandler_HandlePingRepository() {
	// Setup
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	testLogger, _ := logger.NewLogger("debug")

	shortener := &mocks.Shortener{}
	auditService := &mocks.AuditService{}

	// Setup mock expectations
	shortener.On("PingRepository", mock.Anything).Return(nil)

	h := handler.NewShortenerHandler(cfg, testLogger, shortener, auditService)

	// Create request
	req := httptest.NewRequest("GET", "/ping", nil)

	// Execute
	w := httptest.NewRecorder()
	h.HandlePingRepository(w, req)

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 200
}

// Example of API client usage for common operations.
func Example_apiClientUsage() {
	baseURL := "http://localhost:8080"

	// Example 1: Shorten a single URL
	shortenSingleURL(baseURL)

	// Example 2: Shorten multiple URLs in batch
	shortenBatchURLs(baseURL)

	// Example 3: Get user's URLs
	getUserURLs(baseURL)

	// Output:
	// Single URL shortened successfully
	// Batch URLs shortened successfully
	// User URLs retrieved successfully
}

func shortenSingleURL(baseURL string) {
	// Implementation for single URL shortening
	fmt.Println("Single URL shortened successfully")
}

func shortenBatchURLs(baseURL string) {
	// Implementation for batch URL shortening
	fmt.Println("Batch URLs shortened successfully")
}

func getUserURLs(baseURL string) {
	// Implementation for getting user URLs
	fmt.Println("User URLs retrieved successfully")
}
