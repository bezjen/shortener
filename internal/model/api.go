// Package model provides data models and structures for the URL shortening service.
// It defines request/response formats for API endpoints and data transfer objects.
package model

// ShortenJSONRequest represents the JSON request structure for URL shortening endpoint.
// Used in POST /api/shorten endpoint.
//
// Example:
//
//	{
//	  "url": "https://example.com/very-long-url"
//	}
type ShortenJSONRequest struct {
	// URL is the original URL to be shortened.
	// Required: true
	// Example: "https://example.com/very-long-url"
	URL string `json:"url"`
}

// ShortenJSONResponse represents the JSON response structure for URL shortening endpoint.
// Used in POST /api/shorten endpoint responses.
//
// Success Example:
//
//	{
//	  "result": "http://localhost:8080/abc123def"
//	}
//
// Error Example:
//
//	{
//	  "error": "incorrect url"
//	}
type ShortenJSONResponse struct {
	// ShortURL contains the generated short URL on successful shortening.
	// Omitted in error responses.
	// Example: "http://localhost:8080/abc123def"
	ShortURL string `json:"result,omitempty"`

	// Error contains the error message when URL shortening fails.
	// Omitted in successful responses.
	// Example: "incorrect url"
	Error string `json:"error,omitempty"`
}

// ShortenBatchRequestItem represents a single URL item in batch shortening request.
// Used in POST /api/shorten/batch endpoint request body.
//
// Example:
//
//	{
//	  "correlation_id": "request-1",
//	  "original_url": "https://example.com/url1"
//	}
type ShortenBatchRequestItem struct {
	// CorrelationID is a client-provided identifier for matching request and response items.
	// Required: true
	// Example: "request-1"
	CorrelationID string `json:"correlation_id"`

	// OriginalURL is the original URL to be shortened.
	// Required: true
	// Example: "https://example.com/url1"
	OriginalURL string `json:"original_url"`
}

// ShortenBatchResponseItem represents a single shortened URL item in batch response.
// Used in POST /api/shorten/batch endpoint response body.
//
// Example:
//
//	{
//	  "correlation_id": "request-1",
//	  "short_url": "http://localhost:8080/abc123"
//	}
type ShortenBatchResponseItem struct {
	// CorrelationID is the identifier from the corresponding request item.
	// Example: "request-1"
	CorrelationID string `json:"correlation_id"`

	// ShortURL is the generated short URL for the original URL.
	// Example: "http://localhost:8080/abc123"
	ShortURL string `json:"short_url"`
}

// UserURLResponseItem represents a single URL item in user URLs response.
// Used in GET /api/user/urls endpoint response body.
//
// Example:
//
//	{
//	  "short_url": "http://localhost:8080/abc123",
//	  "original_url": "https://example.com/url1"
//	}
type UserURLResponseItem struct {
	// ShortURL is the shortened URL created by the user.
	// Example: "http://localhost:8080/abc123"
	ShortURL string `json:"short_url"`

	// OriginalURL is the original URL that was shortened.
	// Example: "https://example.com/url1"
	OriginalURL string `json:"original_url"`
}

// NewShortenBatchRequestItem creates a new ShortenBatchRequestItem instance.
// Constructor function for batch URL shortening request items.
//
// Parameters:
//   - correlationID: client-provided identifier for request/response matching
//   - OriginalURL: original URL to be shortened
//
// Returns:
//   - *ShortenBatchRequestItem: initialized batch request item
func NewShortenBatchRequestItem(correlationID string, OriginalURL string) *ShortenBatchRequestItem {
	return &ShortenBatchRequestItem{
		CorrelationID: correlationID,
		OriginalURL:   OriginalURL,
	}
}

// NewShortenBatchResponseItem creates a new ShortenBatchResponseItem instance.
// Constructor function for batch URL shortening response items.
//
// Parameters:
//   - correlationID: identifier from the corresponding request item
//   - shortURL: generated short URL
//
// Returns:
//   - *ShortenBatchResponseItem: initialized batch response item
func NewShortenBatchResponseItem(correlationID string, shortURL string) *ShortenBatchResponseItem {
	return &ShortenBatchResponseItem{
		CorrelationID: correlationID,
		ShortURL:      shortURL,
	}
}

// NewUserURLResponseItem creates a new UserURLResponseItem instance.
// Constructor function for user URLs response items.
//
// Parameters:
//   - shortURL: shortened URL created by the user
//   - originalURL: original URL that was shortened
//
// Returns:
//   - *UserURLResponseItem: initialized user URL response item
func NewUserURLResponseItem(shortURL string, originalURL string) *UserURLResponseItem {
	return &UserURLResponseItem{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
