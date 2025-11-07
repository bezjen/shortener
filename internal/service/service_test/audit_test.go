package service_test

import (
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/mocks"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/service"
	"testing"
	"time"
)

func TestShortenerAuditService_RegisterObserver(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	auditService := service.NewShortenerAuditService(testLogger)

	mockObserver := mocks.NewAuditObserver(t)

	auditService.RegisterObserver(mockObserver)

	event := model.NewAuditEvent(time.Now().Unix(), model.ActionShorten, "test-user", "https://example.com")

	mockObserver.On("Notify", *event).Return(nil)

	auditService.NotifyAll(*event)

	mockObserver.AssertCalled(t, "Notify", *event)
}

func TestShortenerAuditService_NotifyAll(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	auditService := service.NewShortenerAuditService(testLogger)

	observer1 := mocks.NewAuditObserver(t)
	observer2 := mocks.NewAuditObserver(t)

	auditService.RegisterObserver(observer1)
	auditService.RegisterObserver(observer2)

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionFollow,
		UserID: "test-user-789",
		URL:    "https://example.com/test/url",
	}

	observer1.On("Notify", event).Return(nil)
	observer2.On("Notify", event).Return(nil)

	auditService.NotifyAll(event)

	observer1.AssertCalled(t, "Notify", event)
	observer2.AssertCalled(t, "Notify", event)
}

func TestShortenerAuditService_NotifyAll_WithError(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	auditService := service.NewShortenerAuditService(testLogger)

	observer1 := mocks.NewAuditObserver(t)
	observer2 := mocks.NewAuditObserver(t)

	auditService.RegisterObserver(observer1)
	auditService.RegisterObserver(observer2)

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	observer1.On("Notify", event).Return(errors.New("mock error"))
	observer2.On("Notify", event).Return(nil)

	auditService.NotifyAll(event)

	observer1.AssertCalled(t, "Notify", event)
	observer2.AssertCalled(t, "Notify", event)
}

func TestShortenerAuditService_MultipleEvents(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	auditService := service.NewShortenerAuditService(testLogger)

	observer := mocks.NewAuditObserver(t)
	auditService.RegisterObserver(observer)

	events := []model.AuditEvent{
		{
			TS:     time.Now().Unix(),
			Action: model.ActionShorten,
			UserID: "user1",
			URL:    "https://example.com/1",
		},
		{
			TS:     time.Now().Unix() + 1,
			Action: model.ActionFollow,
			UserID: "user2",
			URL:    "https://example.com/2",
		},
		{
			TS:     time.Now().Unix() + 2,
			Action: model.ActionShorten,
			UserID: "user3",
			URL:    "https://example.com/3",
		},
	}

	for _, event := range events {
		observer.On("Notify", event).Return(nil)
	}

	for _, event := range events {
		auditService.NotifyAll(event)
	}

	observer.AssertNumberOfCalls(t, "Notify", 3)
	for _, event := range events {
		observer.AssertCalled(t, "Notify", event)
	}
}

func TestShortenerAuditService_ConfigureObservers(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	auditService := service.NewShortenerAuditService(testLogger)

	// Тест с обоими наблюдателями
	cfg := config.Config{
		AuditFile: "/tmp/audit.log",
		AuditURL:  "http://example.com/audit",
	}

	auditService.ConfigureObservers(cfg)

	// Проверяем что наблюдатели зарегистрированы, отправляя событие
	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	// Тест должен завершиться без ошибок
	auditService.NotifyAll(event)

	// Тест только с файловым наблюдателем
	auditService2 := service.NewShortenerAuditService(testLogger)
	cfg2 := config.Config{
		AuditFile: "/tmp/audit.log",
		AuditURL:  "",
	}
	auditService2.ConfigureObservers(cfg2)
	auditService2.NotifyAll(event)

	// Тест только с URL наблюдателем
	auditService3 := service.NewShortenerAuditService(testLogger)
	cfg3 := config.Config{
		AuditFile: "",
		AuditURL:  "http://example.com/audit",
	}
	auditService3.ConfigureObservers(cfg3)
	auditService3.NotifyAll(event)
}

func TestShortenerAuditService_NotifyAll_EmptyObservers(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	auditService := service.NewShortenerAuditService(testLogger)

	// Тест с пустыми наблюдателями
	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	// Не должно быть паники
	auditService.NotifyAll(event)
}
