// Package logger provides logging functionality for the URL shortening service.
// It wraps the zap logging library with a simplified interface.
package logger

import (
	"go.uber.org/zap"
)

// Logger provides a structured logging interface using zap logger.
// It offers leveled logging with sugar methods for convenience.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger creates a new Logger instance with the specified log level.
// Configures production-grade logging with customizable log level.
//
// Parameters:
//   - level: log level string (debug, info, warn, error, etc.)
//
// Returns:
//   - *Logger: initialized logger instance
//   - error: error if log level parsing fails
func NewLogger(level string) (*Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{zl.Sugar()}, nil
}
