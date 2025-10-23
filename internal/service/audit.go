//go:generate mockery --name=AuditObserver --name=AuditService --output=../mocks --case=underscore
package service

import (
	"fmt"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
	"go.uber.org/zap"
)

type AuditObserver interface {
	Notify(event model.AuditEvent) error
}

type AuditService interface {
	RegisterObserver(observer AuditObserver)
	NotifyAll(event model.AuditEvent)
}

type ShortenerAuditService struct {
	observers []AuditObserver
	logger    *logger.Logger
}

func NewShortenerAuditService(logger *logger.Logger) *ShortenerAuditService {
	return &ShortenerAuditService{
		observers: make([]AuditObserver, 0),
		logger:    logger,
	}
}

func (a *ShortenerAuditService) RegisterObserver(observer AuditObserver) {
	a.observers = append(a.observers, observer)
}

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
