package service

import (
	"bytes"
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"net/http"
)

type AuditURL struct {
	url string
}

func NewAuditURL(url string) *AuditURL {
	return &AuditURL{
		url: url,
	}
}

func (a *AuditURL) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(a.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
