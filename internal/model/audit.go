// Package model provides data models and structures for the URL shortening service.
package model

// AuditAction represents the type of action being audited.
type AuditAction string

// Audit constants.
const (
	// ActionShorten represents URL shortening actions.
	// Recorded when a user creates a new short URL.
	ActionShorten AuditAction = "shorten"

	// ActionFollow represents URL access actions.
	// Recorded when a user follows a short URL to access the original URL.
	ActionFollow AuditAction = "follow"
)

// AuditEvent represents an auditable event in the URL shortening service.
// Used for logging user actions for security, analytics, and compliance purposes.
//
// Example:
//
//	{
//	  "ts": 1640995200,
//	  "action": "shorten",
//	  "user_id": "user-123",
//	  "url": "https://example.com/very-long-url"
//	}
type AuditEvent struct {
	// TS is the Unix timestamp when the event occurred.
	// Example: 1640995200
	TS int64 `json:"ts"`

	// Action is the type of action performed (shorten or follow).
	// Example: "shorten"
	Action AuditAction `json:"action"`

	// UserID is the identifier of the user who performed the action.
	// Example: "user-123"
	UserID string `json:"user_id"`

	// URL is the original URL involved in the action.
	// For shorten actions: the URL being shortened.
	// For follow actions: the original URL being accessed.
	// Example: "https://example.com/very-long-url"
	URL string `json:"url"`
}

// NewAuditEvent creates a new AuditEvent instance.
// Constructor function for audit events with timestamp and action details.
//
// Parameters:
//   - ts: Unix timestamp when the event occurred
//   - action: type of action being audited
//   - userID: identifier of the user performing the action
//   - url: URL involved in the action
//
// Returns:
//   - *AuditEvent: initialized audit event
func NewAuditEvent(ts int64, action AuditAction, userID string, url string) *AuditEvent {
	return &AuditEvent{
		TS:     ts,
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
