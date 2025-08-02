package errors

import (
	"context"
	"fmt"
	"time"

	"excel-schema-generator/internal/ports"
)

// ErrorHandler implements the ErrorHandler interface
type ErrorHandler struct {
	logger        ports.LoggingService
	maxRetries    int
	baseDelay     time.Duration
	maxDelay      time.Duration
	retryableErrors map[ErrorType]bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger ports.LoggingService) *ErrorHandler {
	return &ErrorHandler{
		logger:     logger,
		maxRetries: 3,
		baseDelay:  time.Second,
		maxDelay:   30 * time.Second,
		retryableErrors: map[ErrorType]bool{
			FileErrorType:    true,
			NetworkErrorType: true,
			InternalErrorType: false,
			ValidationErrorType: false,
			ConfigErrorType: false,
			ExcelErrorType: false,
			SchemaErrorType: false,
		},
	}
}

// Handle handles an error by logging it and potentially transforming it
func (h *ErrorHandler) Handle(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// Log the error
	h.logError(err)

	// If it's already an AppError, return as-is
	if IsAppError(err) {
		return err
	}

	// Wrap unknown errors as internal errors
	return WrapError(err, InternalErrorType, InternalNilPointerCode, "An unexpected error occurred")
}

// ShouldRetry determines if an operation should be retried based on the error
func (h *ErrorHandler) ShouldRetry(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}

	appErr := GetAppError(err)
	if appErr == nil {
		// Unknown errors are generally not retryable
		return false
	}

	// Check if error type is retryable
	retryable, exists := h.retryableErrors[appErr.Type]
	if !exists {
		return false
	}

	// Check if error is explicitly marked as retryable
	return retryable && appErr.IsRetryable()
}

// GetRetryDelay returns the delay before retrying based on attempt count
func (h *ErrorHandler) GetRetryDelay(ctx context.Context, attempt int) int64 {
	if attempt <= 0 {
		return 0
	}

	// Exponential backoff with jitter
	delay := h.baseDelay * time.Duration(1<<uint(attempt-1))
	if delay > h.maxDelay {
		delay = h.maxDelay
	}

	return int64(delay)
}

// logError logs an error with appropriate level and context
func (h *ErrorHandler) logError(err error) {
	if h.logger == nil {
		return
	}

	appErr := GetAppError(err)
	if appErr == nil {
		h.logger.Error("Unexpected error occurred", "error", err.Error())
		return
	}

	// Prepare log context
	logArgs := []interface{}{
		"error_type", appErr.Type,
		"error_code", appErr.Code,
		"error", err.Error(),
	}

	// Add context if available
	if appErr.Context != nil && len(appErr.Context) > 0 {
		logArgs = append(logArgs, "context", appErr.Context)
	}

	// Add cause if available
	if appErr.Cause != nil {
		logArgs = append(logArgs, "cause", appErr.Cause.Error())
	}

	// Log with appropriate level based on error type
	switch appErr.Type {
	case ValidationErrorType:
		h.logger.Warn("Validation error", logArgs...)
	case FileErrorType, ExcelErrorType:
		h.logger.Error("Processing error", logArgs...)
	case ConfigErrorType:
		h.logger.Error("Configuration error", logArgs...)
	case InternalErrorType:
		h.logger.Error("Internal error", logArgs...)
	default:
		h.logger.Error("Application error", logArgs...)
	}
}

// RecoverFromPanic recovers from a panic and converts it to an error
func (h *ErrorHandler) RecoverFromPanic() error {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic recovered: %v", r)
		h.logger.Error("Panic recovered", "panic", r)
		return WrapError(err, InternalErrorType, InternalStateInconsistentCode, "A critical error occurred")
	}
	return nil
}

// SetRetryConfig sets retry configuration
func (h *ErrorHandler) SetRetryConfig(maxRetries int, baseDelay, maxDelay time.Duration) {
	h.maxRetries = maxRetries
	h.baseDelay = baseDelay
	h.maxDelay = maxDelay
}

// SetRetryable marks an error type as retryable or not
func (h *ErrorHandler) SetRetryable(errorType ErrorType, retryable bool) {
	h.retryableErrors[errorType] = retryable
}

// FormatUserFriendlyMessage formats an error message for end users
func FormatUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	appErr := GetAppError(err)
	if appErr == nil {
		return "An unexpected error occurred. Please try again."
	}

	switch appErr.Type {
	case ValidationErrorType:
		return fmt.Sprintf("Validation error: %s", appErr.Message)
	case FileErrorType:
		switch appErr.Code {
		case FileNotFoundCode:
			return "The specified file could not be found. Please check the file path and try again."
		case FilePermissionCode:
			return "Permission denied. Please check that you have the necessary permissions to access the file."
		case FileCorruptedCode:
			return "The file appears to be corrupted or damaged. Please try with a different file."
		default:
			return fmt.Sprintf("File error: %s", appErr.Message)
		}
	case ExcelErrorType:
		switch appErr.Code {
		case ExcelInvalidFormatCode:
			return "The file is not a valid Excel file. Please ensure you're using a .xlsx or .xls file."
		case ExcelPasswordProtectedCode:
			return "The Excel file is password protected. Please provide an unprotected file."
		case ExcelSheetNotFoundCode:
			return "The specified sheet could not be found in the Excel file."
		default:
			return fmt.Sprintf("Excel processing error: %s", appErr.Message)
		}
	case SchemaErrorType:
		return fmt.Sprintf("Schema error: %s", appErr.Message)
	case ConfigErrorType:
		return fmt.Sprintf("Configuration error: %s", appErr.Message)
	default:
		return "An error occurred while processing your request. Please try again."
	}
}