package errors

import (
	"context"
	"errors"
	"sync"
	"testing"

	"excel-schema-generator/internal/ports"
)

// Mock logger for testing
type mockLogger struct {
	mu         sync.RWMutex
	debugCalls []logCall
	infoCalls  []logCall
	warnCalls  []logCall
	errorCalls []logCall
}

type logCall struct {
	msg  string
	args []any
}

func (m *mockLogger) Debug(msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugCalls = append(m.debugCalls, logCall{msg: msg, args: keysAndValues})
}

func (m *mockLogger) Info(msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoCalls = append(m.infoCalls, logCall{msg: msg, args: keysAndValues})
}

func (m *mockLogger) Warn(msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnCalls = append(m.warnCalls, logCall{msg: msg, args: keysAndValues})
}

func (m *mockLogger) Error(msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalls = append(m.errorCalls, logCall{msg: msg, args: keysAndValues})
}

func (m *mockLogger) With(keysAndValues ...any) ports.LoggingService {
	return m
}

func TestErrorHandler_HandleValidationError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	err := NewValidationError(ValidationRequiredFieldCode, "Field is required")
	
	result := handler.Handle(ctx, err)
	
	if result == nil {
		t.Fatal("Expected error result, got nil")
	}
	
	appErr, ok := result.(*AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", result)
	}
	
	if appErr.Code != ValidationRequiredFieldCode {
		t.Errorf("Expected code %s, got %s", ValidationRequiredFieldCode, appErr.Code)
	}
	
	if appErr.Message != "Field is required" {
		t.Errorf("Expected message 'Field is required', got '%s'", appErr.Message)
	}
	
	if len(logger.warnCalls) != 1 {
		t.Errorf("Expected 1 warn log call for validation error, got %d", len(logger.warnCalls))
	}
}

func TestErrorHandler_HandleFileError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	err := NewFileError(FileNotFoundCode, "File not found").WithContext("file", "test.xlsx")
	
	result := handler.Handle(ctx, err)
	
	if result == nil {
		t.Fatal("Expected error result, got nil")
	}
	
	appErr, ok := result.(*AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", result)
	}
	
	if appErr.Code != FileNotFoundCode {
		t.Errorf("Expected code %s, got %s", FileNotFoundCode, appErr.Code)
	}
	
	if appErr.Context["file"] != "test.xlsx" {
		t.Errorf("Expected file context 'test.xlsx', got '%v'", appErr.Context["file"])
	}
	
	if len(logger.errorCalls) != 1 {
		t.Errorf("Expected 1 error log call, got %d", len(logger.errorCalls))
	}
}

func TestErrorHandler_HandleExcelError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	err := NewExcelError(ExcelInvalidFormatCode, "Invalid Excel format").
		WithContext("file", "test.xlsx").
		WithContext("sheet", "Sheet1")
	
	result := handler.Handle(ctx, err)
	
	if result == nil {
		t.Fatal("Expected error result, got nil")
	}
	
	appErr, ok := result.(*AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", result)
	}
	
	if appErr.Code != ExcelInvalidFormatCode {
		t.Errorf("Expected code %s, got %s", ExcelInvalidFormatCode, appErr.Code)
	}
	
	if appErr.Context["file"] != "test.xlsx" {
		t.Errorf("Expected file context 'test.xlsx', got '%v'", appErr.Context["file"])
	}
	
	if appErr.Context["sheet"] != "Sheet1" {
		t.Errorf("Expected sheet context 'Sheet1', got '%v'", appErr.Context["sheet"])
	}
}

func TestErrorHandler_HandleSchemaError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	err := NewSchemaError(SchemaValidationFailedCode, "Schema validation failed")
	
	result := handler.Handle(ctx, err)
	
	if result == nil {
		t.Fatal("Expected error result, got nil")
	}
	
	appErr, ok := result.(*AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", result)
	}
	
	if appErr.Code != SchemaValidationFailedCode {
		t.Errorf("Expected code %s, got %s", SchemaValidationFailedCode, appErr.Code)
	}
}

func TestErrorHandler_HandleGenericError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	err := errors.New("generic error")
	
	result := handler.Handle(ctx, err)
	
	if result == nil {
		t.Fatal("Expected error result, got nil")
	}
	
	appErr, ok := result.(*AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", result)
	}
	
	if appErr.Type != InternalErrorType {
		t.Errorf("Expected type %s, got %s", InternalErrorType, appErr.Type)
	}
	
	if appErr.Cause == nil {
		t.Error("Expected cause to be set for wrapped error")
	}
	
	if len(logger.errorCalls) != 1 {
		t.Errorf("Expected 1 error log call, got %d", len(logger.errorCalls))
	}
}

func TestErrorHandler_HandleNilError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	result := handler.Handle(ctx, nil)
	
	if result != nil {
		t.Errorf("Expected nil result for nil error, got %v", result)
	}
	
	if len(logger.errorCalls) != 0 {
		t.Errorf("Expected 0 error log calls for nil error, got %d", len(logger.errorCalls))
	}
}

func TestErrorHandler_HandleAppError(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	originalErr := &AppError{
		Type:    ValidationErrorType,
		Code:    ValidationRequiredFieldCode,
		Message: "Test error",
		Context: map[string]interface{}{"field": "test"},
		Cause:   errors.New("root cause"),
	}
	
	result := handler.Handle(ctx, originalErr)
	
	if result != originalErr {
		t.Error("Expected same AppError instance to be returned")
	}
	
	if len(logger.warnCalls) != 1 {
		t.Errorf("Expected 1 warn log call for validation error, got %d", len(logger.warnCalls))
	}
}

func TestFormatUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "Validation error",
			err: &AppError{
				Type:    ValidationErrorType,
				Code:    ValidationRequiredFieldCode,
				Message: "Field is required",
			},
			expected: "Validation error: Field is required",
		},
		{
			name: "File not found error",
			err: &AppError{
				Type:    FileErrorType,
				Code:    FileNotFoundCode,
				Message: "File not found",
				Context: map[string]interface{}{"file": "test.xlsx"},
			},
			expected: "The specified file could not be found. Please check the file path and try again.",
		},
		{
			name: "Excel processing error",
			err: &AppError{
				Type:    ExcelErrorType,
				Code:    ExcelInvalidFormatCode,
				Message: "Invalid format",
				Context: map[string]interface{}{"file": "test.xlsx", "sheet": "Sheet1"},
			},
			expected: "The file is not a valid Excel file. Please ensure you're using a .xlsx or .xls file.",
		},
		{
			name: "Generic error without context",
			err: &AppError{
				Type:    InternalErrorType,
				Code:    InternalNilPointerCode,
				Message: "Internal error",
			},
			expected: "An error occurred while processing your request. Please try again.",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatUserFriendlyMessage(tt.err)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestErrorHandler_Concurrent(t *testing.T) {
	logger := &mockLogger{}
	handler := NewErrorHandler(logger)
	ctx := context.Background()
	
	// Test concurrent error handling
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := NewValidationError(ValidationRequiredFieldCode, "Concurrent test error")
			result := handler.Handle(ctx, err)
			
			if result == nil {
				t.Errorf("Expected error result for goroutine %d, got nil", id)
			}
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	logger.mu.RLock()
	warnCallCount := len(logger.warnCalls)
	logger.mu.RUnlock()
	
	if warnCallCount != 10 {
		t.Errorf("Expected 10 warn log calls for validation errors, got %d", warnCallCount)
	}
}

func TestAppError_Methods(t *testing.T) {
	err := NewValidationError(ValidationRequiredFieldCode, "Test error")
	
	// Test WithContext
	err = err.WithContext("field", "test_field")
	if err.Context["field"] != "test_field" {
		t.Error("WithContext should add context")
	}
	
	// Test WithCause
	rootCause := errors.New("root cause")
	err = err.WithCause(rootCause)
	if err.Cause != rootCause {
		t.Error("WithCause should set the cause")
	}
	
	// Test Unwrap
	if err.Unwrap() != rootCause {
		t.Error("Unwrap should return the cause")
	}
	
	// Test IsRetryable (default false)
	if err.IsRetryable() {
		t.Error("Default error should not be retryable")
	}
	
	// Make it retryable
	err.Retryable = true
	if !err.IsRetryable() {
		t.Error("Error should be retryable after setting flag")
	}
}

func TestAppError_Wrapping(t *testing.T) {
	rootCause := errors.New("original error")
	wrappedErr := WrapError(rootCause, ValidationErrorType, ValidationInvalidValueCode, "Validation failed")
	
	if wrappedErr.Cause != rootCause {
		t.Error("WrapError should preserve the original error")
	}
	
	if wrappedErr.Type != ValidationErrorType {
		t.Error("WrapError should set the correct type")
	}
	
	if wrappedErr.Code != ValidationInvalidValueCode {
		t.Error("WrapError should set the correct code")
	}
	
	// Test error chain
	if !errors.Is(wrappedErr, rootCause) {
		t.Error("Wrapped error should be identifiable as root cause")
	}
}

func TestIsAppError(t *testing.T) {
	appErr := NewValidationError(ValidationRequiredFieldCode, "Test error")
	genericErr := errors.New("generic error")
	
	if !IsAppError(appErr) {
		t.Error("IsAppError should return true for AppError")
	}
	
	if IsAppError(genericErr) {
		t.Error("IsAppError should return false for generic error")
	}
	
	if IsAppError(nil) {
		t.Error("IsAppError should return false for nil")
	}
}

func TestGetAppError(t *testing.T) {
	appErr := NewValidationError(ValidationRequiredFieldCode, "Test error")
	genericErr := errors.New("generic error")
	
	result := GetAppError(appErr)
	if result != appErr {
		t.Error("GetAppError should return the same AppError")
	}
	
	result = GetAppError(genericErr)
	if result != nil {
		t.Error("GetAppError should return nil for non-AppError")
	}
	
	result = GetAppError(nil)
	if result != nil {
		t.Error("GetAppError should return nil for nil error")
	}
}