package errors

import (
	"fmt"
)

// ErrorType represents different types of errors in the application
type ErrorType string

const (
	// ValidationError represents validation errors
	ValidationErrorType ErrorType = "VALIDATION_ERROR"
	
	// FileError represents file system errors
	FileErrorType ErrorType = "FILE_ERROR"
	
	// ExcelError represents Excel processing errors
	ExcelErrorType ErrorType = "EXCEL_ERROR"
	
	// SchemaError represents schema-related errors
	SchemaErrorType ErrorType = "SCHEMA_ERROR"
	
	// ConfigError represents configuration errors
	ConfigErrorType ErrorType = "CONFIG_ERROR"
	
	// InternalError represents internal application errors
	InternalErrorType ErrorType = "INTERNAL_ERROR"
	
	// NetworkError represents network-related errors
	NetworkErrorType ErrorType = "NETWORK_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	Cause      error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Retryable  bool                   `json:"retryable"`
	StatusCode int                    `json:"status_code,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// IsRetryable returns whether the error is retryable
func (e *AppError) IsRetryable() bool {
	return e.Retryable
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string) *AppError {
	return NewAppError(ValidationErrorType, code, message)
}

// NewFileError creates a new file error
func NewFileError(code, message string) *AppError {
	return NewAppError(FileErrorType, code, message)
}

// NewExcelError creates a new Excel processing error
func NewExcelError(code, message string) *AppError {
	return NewAppError(ExcelErrorType, code, message)
}

// NewSchemaError creates a new schema error
func NewSchemaError(code, message string) *AppError {
	return NewAppError(SchemaErrorType, code, message)
}

// NewConfigError creates a new configuration error
func NewConfigError(code, message string) *AppError {
	return NewAppError(ConfigErrorType, code, message)
}

// NewInternalError creates a new internal error
func NewInternalError(code, message string) *AppError {
	return NewAppError(InternalErrorType, code, message)
}

// WrapError wraps an existing error with application error context
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	return NewAppError(errorType, code, message).WithCause(err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error, returns nil if not an AppError
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// ErrorCode constants for common errors
const (
	// File operation error codes
	FileNotFoundCode     = "FILE_NOT_FOUND"
	FilePermissionCode   = "FILE_PERMISSION"
	FileCorruptedCode    = "FILE_CORRUPTED"
	DirectoryNotFoundCode = "DIRECTORY_NOT_FOUND"
	
	// Excel processing error codes
	ExcelInvalidFormatCode = "EXCEL_INVALID_FORMAT"
	ExcelCorruptedCode     = "EXCEL_CORRUPTED"
	ExcelPasswordProtectedCode = "EXCEL_PASSWORD_PROTECTED"
	ExcelSheetNotFoundCode = "EXCEL_SHEET_NOT_FOUND"
	
	// Schema error codes
	SchemaInvalidCode      = "SCHEMA_INVALID"
	SchemaVersionMismatchCode = "SCHEMA_VERSION_MISMATCH"
	SchemaMissingFieldCode = "SCHEMA_MISSING_FIELD"
	SchemaValidationFailedCode = "SCHEMA_VALIDATION_FAILED"
	
	// Validation error codes
	ValidationRequiredFieldCode = "VALIDATION_REQUIRED_FIELD"
	ValidationInvalidTypeCode   = "VALIDATION_INVALID_TYPE"
	ValidationInvalidValueCode  = "VALIDATION_INVALID_VALUE"
	ValidationConstraintCode    = "VALIDATION_CONSTRAINT"
	
	// Configuration error codes
	ConfigMissingCode    = "CONFIG_MISSING"
	ConfigInvalidCode    = "CONFIG_INVALID"
	ConfigParseFailedCode = "CONFIG_PARSE_FAILED"
	
	// Internal error codes
	InternalNilPointerCode     = "INTERNAL_NIL_POINTER"
	InternalStateInconsistentCode = "INTERNAL_STATE_INCONSISTENT"
	InternalResourceExhaustedCode = "INTERNAL_RESOURCE_EXHAUSTED"
)