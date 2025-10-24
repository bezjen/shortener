// Package service provides business logic for URL shortening service.
package service

import (
	"bytes"
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"net/http"
)

// AuditURL implements AuditObserver for sending audit events to a remote HTTP endpoint.
type AuditURL struct {
	url string
}

// NewAuditURL creates a new AuditURL observer for remote audit logging.
//
// Parameters:
//   - url: HTTP endpoint URL to send audit events to
//
// Returns:
//   - *AuditURL: Initialized remote audit observer
func NewAuditURL(url string) *AuditURL {
	return &AuditURL{
		url: url,
	}
}

// Notify sends an audit event to the configured remote URL.
// The event is serialized as JSON and sent via HTTP POST request.
//
// Parameters:
//   - event: Audit event to be sent
//
// Returns:
//   - error: Error if HTTP request fails or event serialization fails
func (a *AuditURL) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(a.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
