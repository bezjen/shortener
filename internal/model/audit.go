package model

type AuditAction string

const (
	ActionShorten AuditAction = "shorten"
	ActionFollow  AuditAction = "follow"
)

type AuditEvent struct {
	TS     int64       `json:"ts"`
	Action AuditAction `json:"action"`
	UserID string      `json:"user_id"`
	URL    string      `json:"url"`
}

func NewAuditEvent(ts int64, action AuditAction, userID string, url string) *AuditEvent {
	return &AuditEvent{
		TS:     ts,
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
