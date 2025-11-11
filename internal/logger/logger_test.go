package logger

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{
			name:      "valid debug level",
			level:     "debug",
			wantError: false,
		},
		{
			name:      "valid info level",
			level:     "info",
			wantError: false,
		},
		{
			name:      "invalid level",
			level:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level)
			if (err != nil) != tt.wantError {
				t.Errorf("NewLogger() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && logger == nil {
				t.Error("Expected logger instance, got nil")
			}
		})
	}
}
