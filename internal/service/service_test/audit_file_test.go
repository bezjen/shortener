package service_test

import (
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
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
