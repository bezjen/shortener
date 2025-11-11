package service_test

import (
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAuditFile_Notify(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditFile := service.NewAuditFile(tmpFile.Name())

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user-123",
		URL:    "https://example.com/long/url",
	}

	err = auditFile.Notify(event)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var writtenEvent model.AuditEvent
	err = json.Unmarshal(data, &writtenEvent)
	require.NoError(t, err)

	assert.Equal(t, event.TS, writtenEvent.TS)
	assert.Equal(t, event.Action, writtenEvent.Action)
	assert.Equal(t, event.UserID, writtenEvent.UserID)
	assert.Equal(t, event.URL, writtenEvent.URL)
}

func TestAuditFile_Notify_InvalidPath(t *testing.T) {
	auditFile := service.NewAuditFile("/invalid/pathaudit.log")

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	err := auditFile.Notify(event)
	assert.Error(t, err)
}

func TestAuditFile_Notify_Concurrent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_concurrent_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditFile := service.NewAuditFile(tmpFile.Name())

	// Запускаем несколько горутин для проверки конкурентности
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			event := model.AuditEvent{
				TS:     time.Now().Unix() + int64(index),
				Action: model.ActionShorten,
				UserID: "test-user",
				URL:    "https://example.com",
			}
			err := auditFile.Notify(event)
			assert.NoError(t, err)
		}(i)
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем что все события записаны
	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	// Разделяем на строки и подсчитываем
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Equal(t, 10, len(lines))

	// Проверяем, что каждая строка является валидным JSON
	for _, line := range lines {
		var event model.AuditEvent
		err = json.Unmarshal([]byte(line), &event)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", event.UserID)
		assert.Equal(t, model.ActionShorten, event.Action)
	}
}
