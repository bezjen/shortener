package service

import (
	"encoding/json"
	"github.com/bezjen/shortener/internal/model"
	"os"
	"sync"
)

type AuditFile struct {
	filePath string
	mu       sync.Mutex
}

func NewAuditFile(filePath string) *AuditFile {
	return &AuditFile{
		filePath: filePath,
	}
}

func (a *AuditFile) Notify(event model.AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	return err
}
