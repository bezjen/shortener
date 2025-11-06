// Package service provides business logic for URL shortening service.
//
//go:generate mockery --name=AuditObserver --name=AuditService --output=../mocks --case=underscore
package service

import (
	"fmt"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
	"go.uber.org/zap"
)

// AuditObserver defines the interface for audit event observers.
type AuditObserver interface {
	Notify(event model.AuditEvent) error
}

// AuditService manages audit observers and distributes events to them.
type AuditService interface {
	RegisterObserver(observer AuditObserver)
	NotifyAll(event model.AuditEvent)
}

// ShortenerAuditService implements AuditService for URL shortening operations.
// It manages a collection of observers and coordinates event distribution.
type ShortenerAuditService struct {
	observers []AuditObserver
	logger    *logger.Logger
}

// NewShortenerAuditService creates a new ShortenerAuditService instance.
//
// Parameters:
//   - logger: Logger instance for error logging
//
// Returns:
//   - *ShortenerAuditService: Initialized audit service
func NewShortenerAuditService(logger *logger.Logger) *ShortenerAuditService {
	return &ShortenerAuditService{
		observers: make([]AuditObserver, 0),
		logger:    logger,
	}
}

// ConfigureObservers configures audit observers based on application configuration.
// Sets up file and/or remote URL observers if configured.
//
// Parameters:
//   - cfg: Application configuration containing audit settings
func (a *ShortenerAuditService) ConfigureObservers(cfg config.Config) {
	if cfg.AuditFile != "" {
		fileAudit := NewAuditFile(cfg.AuditFile)
		a.RegisterObserver(fileAudit)
	}
	if cfg.AuditURL != "" {
		remoteAudit := NewAuditURL(cfg.AuditURL)
		a.RegisterObserver(remoteAudit)
	}
}

// RegisterObserver adds a new observer to the audit service.
// The observer will receive all future audit events.
//
// Parameters:
//   - observer: Observer implementation to be registered
func (a *ShortenerAuditService) RegisterObserver(observer AuditObserver) {
	a.observers = append(a.observers, observer)
}

// NotifyAll distributes an audit event to all registered observers.
// If an observer fails to process the event, it's logged but doesn't stop other observers.
//
// Parameters:
//   - event: Audit event to be distributed
func (a *ShortenerAuditService) NotifyAll(event model.AuditEvent) {
	for _, observer := range a.observers {
		a.notify(observer, event)
	}
}

func (a *ShortenerAuditService) notify(observer AuditObserver, event model.AuditEvent) {
	if err := observer.Notify(event); err != nil {
		a.logger.Error("Failed to notify",
			zap.String("event", fmt.Sprintf("%v", event)),
			zap.Error(err))
	}
}
