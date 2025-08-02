package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"excel-schema-generator/pkg/logger"
)

func TestLoggerAdapter_Debug(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelDebug,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	adapter.Debug("test debug message", "key", "value")
	
	output := buf.String()
	
	if !strings.Contains(output, "test debug message") {
		t.Errorf("Expected debug message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected key-value pair not found in output: %s", output)
	}
}

func TestLoggerAdapter_Info(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	adapter.Info("test info message", "key1", "value1", "key2", "value2")
	
	output := buf.String()
	
	if !strings.Contains(output, "test info message") {
		t.Errorf("Expected info message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Expected first key-value pair not found in output: %s", output)
	}
	
	if !strings.Contains(output, "key2=value2") {
		t.Errorf("Expected second key-value pair not found in output: %s", output)
	}
}

func TestLoggerAdapter_Warn(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelWarn,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	adapter.Warn("test warning message", "reason", "test case")
	
	output := buf.String()
	
	if !strings.Contains(output, "WARN") {
		t.Error("Warning level not found in output")
	}
	
	if !strings.Contains(output, "test warning message") {
		t.Errorf("Expected warning message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "reason=\"test case\"") {
		t.Errorf("Expected key-value pair not found in output: %s", output)
	}
}

func TestLoggerAdapter_Error(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelError,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	adapter.Error("test error message", "error", "test error")
	
	output := buf.String()
	
	if !strings.Contains(output, "ERROR") {
		t.Error("Error level not found in output")
	}
	
	if !strings.Contains(output, "test error message") {
		t.Errorf("Expected error message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "error=\"test error\"") {
		t.Errorf("Expected key-value pair not found in output: %s", output)
	}
}

func TestLoggerAdapter_With(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	// Create a new logger with context
	contextLogger := adapter.With("component", "test", "session", "123")
	
	// Log with the context logger
	contextLogger.Info("test message with context")
	
	output := buf.String()
	
	if !strings.Contains(output, "test message with context") {
		t.Errorf("Expected message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "component=test") {
		t.Errorf("Expected context key-value pair not found in output: %s", output)
	}
	
	if !strings.Contains(output, "session=123") {
		t.Errorf("Expected context key-value pair not found in output: %s", output)
	}
}

func TestLoggerAdapter_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	adapter.Info("json test message", "key", "value")
	
	output := buf.String()
	
	if !strings.Contains(output, `"msg":"json test message"`) {
		t.Errorf("Expected JSON message not found in output: %s", output)
	}
	
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected JSON key-value pair not found in output: %s", output)
	}
}

func TestLoggerAdapter_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelWarn,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	// These should be filtered out
	adapter.Debug("debug message")
	adapter.Info("info message")
	
	// These should appear
	adapter.Warn("warn message")
	adapter.Error("error message")
	
	output := buf.String()
	
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear")
	}
	
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear")
	}
}

func TestLoggerAdapter_TypeAssertions(t *testing.T) {
	baseLogger := logger.New(logger.DefaultConfig())
	
	// Test that NewLoggerAdapter returns the correct interface
	loggingService := NewLoggerAdapter(baseLogger)
	
	// Test that we can get a concrete type if needed
	adapter, ok := loggingService.(*LoggerAdapter)
	if !ok {
		t.Error("NewLoggerAdapter should return *LoggerAdapter")
	}
	
	// Test that the adapter has the expected fields
	if adapter.logger == nil {
		t.Error("LoggerAdapter.logger should not be nil")
	}
}

func TestLoggerAdapter_WithChaining(t *testing.T) {
	var buf bytes.Buffer
	
	config := logger.Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	baseLogger := logger.New(config)
	adapter := NewLoggerAdapter(baseLogger)
	
	// Chain multiple With calls
	contextLogger := adapter.With("service", "test").With("request", "abc123")
	
	contextLogger.Info("chained context test")
	
	output := buf.String()
	
	if !strings.Contains(output, "chained context test") {
		t.Errorf("Expected message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "service=test") {
		t.Errorf("Expected first context not found in output: %s", output)
	}
	
	if !strings.Contains(output, "request=abc123") {
		t.Errorf("Expected second context not found in output: %s", output)
	}
}