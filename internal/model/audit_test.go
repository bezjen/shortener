package model

import (
	"testing"
	"time"
)

func TestNewAuditEvent(t *testing.T) {
	ts := time.Now().Unix()
	action := ActionShorten
	userID := "user-123"
	url := "https://example.com"

	event := NewAuditEvent(ts, action, userID, url)

	if event.TS != ts {
		t.Errorf("Expected TS %d, got %d", ts, event.TS)
	}

	if event.Action != action {
		t.Errorf("Expected Action %s, got %s", action, event.Action)
	}

	if event.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, event.UserID)
	}

	if event.URL != url {
		t.Errorf("Expected URL %s, got %s", url, event.URL)
	}
}

func TestAuditActionConstants(t *testing.T) {
	if ActionShorten != "shorten" {
		t.Errorf("Expected ActionShorten to be 'shorten', got %s", ActionShorten)
	}

	if ActionFollow != "follow" {
		t.Errorf("Expected ActionFollow to be 'follow', got %s", ActionFollow)
	}
}
