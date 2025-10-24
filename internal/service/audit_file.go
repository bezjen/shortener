// Package service provides business logic for URL shortening service.
package service

import (
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"os"
	"sync"
)

// AuditFile implements AuditObserver for writing audit events to a local file.
type AuditFile struct {
	filePath string
	mu       sync.Mutex
}

// NewAuditFile creates a new AuditFile observer for file-based audit logging.
//
// Parameters:
//   - filePath: Path to the audit log file
//
// Returns:
//   - *AuditFile: Initialized file audit observer
func NewAuditFile(filePath string) *AuditFile {
	return &AuditFile{
		filePath: filePath,
	}
}

// Notify writes an audit event to the audit file as a JSON line.
// The method uses mutex locking to ensure thread-safe file access.
//
// Parameters:
//   - event: Audit event to be written
//
// Returns:
//   - error: Error if file operations fail or JSON serialization fails
func (a *AuditFile) Notify(event model.AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	return err
}
