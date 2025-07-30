package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		Level:  slog.LevelDebug,
		Format: "text",
		Output: &buf,
	}
	
	logger := New(config)
	logger.Debug("test debug message", "key", "value")
	
	output := buf.String()
	
	if !strings.Contains(output, "test debug message") {
		t.Errorf("Expected log message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected key-value pair not found in output: %s", output)
	}
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}
	
	logger := New(config)
	logger.Info("test json message", "key", "value")
	
	output := buf.String()
	
	if !strings.Contains(output, `"msg":"test json message"`) {
		t.Errorf("Expected JSON message not found in output: %s", output)
	}
	
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected JSON key-value pair not found in output: %s", output)
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		Level:  slog.LevelWarn,
		Format: "text",
		Output: &buf,
	}
	
	logger := New(config)
	
	// Debug and Info should be filtered out
	logger.Debug("debug message")
	logger.Info("info message")
	
	// Warn and Error should appear
	logger.Warn("warn message")
	logger.Error("error message")
	
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

func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	
	// Create a new logger and set it as default
	config := Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	newLogger := New(config)
	SetDefault(newLogger)
	
	// Test convenience functions
	Info("test default logger", "test", "value")
	
	output := buf.String()
	
	if !strings.Contains(output, "test default logger") {
		t.Errorf("Expected message not found in output: %s", output)
	}
	
	if !strings.Contains(output, "test=value") {
		t.Errorf("Expected key-value pair not found in output: %s", output)
	}
}

func TestProgress(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	logger := New(config)
	SetDefault(logger)
	
	Progress("Processing files", 5, 10, "file", "test.xlsx")
	
	output := buf.String()
	
	if !strings.Contains(output, "Processing files") {
		t.Error("Progress message not found")
	}
	
	if !strings.Contains(output, "current=5") {
		t.Error("Current value not found")
	}
	
	if !strings.Contains(output, "total=10") {
		t.Error("Total value not found")
	}
	
	if !strings.Contains(output, "file=test.xlsx") {
		t.Error("Additional argument not found")
	}
}

func TestFileProcessed(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	logger := New(config)
	SetDefault(logger)
	
	FileProcessed("test.xlsx", "success", "100ms")
	
	output := buf.String()
	
	if !strings.Contains(output, "File processed") {
		t.Error("File processed message not found")
	}
	
	if !strings.Contains(output, "file=test.xlsx") {
		t.Error("Filename not found")
	}
	
	if !strings.Contains(output, "status=success") {
		t.Error("Status not found")
	}
	
	if !strings.Contains(output, "duration=100ms") {
		t.Error("Duration not found")
	}
}

func TestBatchProgress(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}
	
	logger := New(config)
	SetDefault(logger)
	
	BatchProgress("generate", 3, 10, "test.xlsx")
	
	output := buf.String()
	
	if !strings.Contains(output, "Batch progress") {
		t.Error("Batch progress message not found")
	}
	
	if !strings.Contains(output, "operation=generate") {
		t.Error("Operation not found")
	}
	
	if !strings.Contains(output, "current=3") {
		t.Error("Current value not found")
	}
	
	if !strings.Contains(output, "total=10") {
		t.Error("Total value not found")
	}
	
	if !strings.Contains(output, "current_file=test.xlsx") {
		t.Error("Current file not found")
	}
}