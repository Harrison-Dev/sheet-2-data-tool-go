package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// Config represents logger configuration
type Config struct {
	Level  slog.Level
	Format string // "text" or "json"
	Output io.Writer
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: os.Stdout,
	}
}

// New creates a new logger with the given configuration
func New(config Config) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: config.Level,
	}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(config.Output, opts)
	default:
		handler = slog.NewTextHandler(config.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Global logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = New(DefaultConfig())
}

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// GetDefault returns the default logger
func GetDefault() *Logger {
	return defaultLogger
}

// Convenience functions that use the default logger
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Progress logs progress information
func Progress(msg string, current, total int, args ...any) {
	allArgs := append([]any{"current", current, "total", total}, args...)
	defaultLogger.Info(msg, allArgs...)
}

// FileProcessed logs when a file has been processed
func FileProcessed(filename string, status string, duration string) {
	defaultLogger.Info("File processed",
		"file", filename,
		"status", status,
		"duration", duration,
	)
}

// BatchProgress logs batch processing progress
func BatchProgress(operation string, current, total int, currentFile string) {
	defaultLogger.Info("Batch progress",
		"operation", operation,
		"current", current,
		"total", total,
		"current_file", currentFile,
	)
}