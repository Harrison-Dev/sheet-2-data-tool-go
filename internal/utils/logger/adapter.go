package logger

import (
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/pkg/logger"
)

// LoggerAdapter adapts the pkg/logger.Logger to implement ports.LoggingService
type LoggerAdapter struct {
	logger *logger.Logger
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(logger *logger.Logger) ports.LoggingService {
	return &LoggerAdapter{
		logger: logger,
	}
}

// Debug logs a debug message
func (a *LoggerAdapter) Debug(msg string, keysAndValues ...any) {
	a.logger.Debug(msg, keysAndValues...)
}

// Info logs an info message
func (a *LoggerAdapter) Info(msg string, keysAndValues ...any) {
	a.logger.Info(msg, keysAndValues...)
}

// Warn logs a warning message
func (a *LoggerAdapter) Warn(msg string, keysAndValues ...any) {
	a.logger.Warn(msg, keysAndValues...)
}

// Error logs an error message
func (a *LoggerAdapter) Error(msg string, keysAndValues ...any) {
	a.logger.Error(msg, keysAndValues...)
}

// With returns a new logger with additional context
func (a *LoggerAdapter) With(keysAndValues ...any) ports.LoggingService {
	newLogger := &logger.Logger{Logger: a.logger.With(keysAndValues...)}
	return NewLoggerAdapter(newLogger)
}