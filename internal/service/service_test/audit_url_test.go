package service_test

import (
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuditURL_Notify_Success(t *testing.T) {
	var requestCount int32
	var lastEvent model.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var event model.AuditEvent
		err := json.NewDecoder(r.Body).Decode(&event)
		require.NoError(t, err)
		lastEvent = event
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	auditURL := service.NewAuditURL(server.URL)

	expectedEvent := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionFollow,
		UserID: "test-user-456",
		URL:    "https://example.com/another/url",
	}

	err := auditURL.Notify(expectedEvent)
	require.NoError(t, err)

	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))

	assert.Equal(t, expectedEvent.TS, lastEvent.TS)
	assert.Equal(t, expectedEvent.Action, lastEvent.Action)
	assert.Equal(t, expectedEvent.UserID, lastEvent.UserID)
	assert.Equal(t, expectedEvent.URL, lastEvent.URL)
}

func TestAuditURL_Notify_InvalidURL(t *testing.T) {
	auditURL := service.NewAuditURL("invalid-url")

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	err := auditURL.Notify(event)
	assert.Error(t, err)
}

func TestAuditURL_Notify_ServerError(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	auditURL := service.NewAuditURL(server.URL)

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	err := auditURL.Notify(event)
	require.NoError(t, err) // AuditURL не проверяет статус ответа

	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestAuditURL_Notify_Timeout(t *testing.T) {
	// Сервер который долго отвечает
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	auditURL := service.NewAuditURL(server.URL)

	event := model.AuditEvent{
		TS:     time.Now().Unix(),
		Action: model.ActionShorten,
		UserID: "test-user",
		URL:    "https://example.com",
	}

	err := auditURL.Notify(event)
	assert.NoError(t, err)
}
