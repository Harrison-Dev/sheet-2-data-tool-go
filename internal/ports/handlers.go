package ports

import (
	"context"

	"excel-schema-generator/internal/core/models"
)

// CommandHandler defines the interface for handling command operations
type CommandHandler interface {
	// Handle processes a command and returns a result
	Handle(ctx context.Context, cmd Command) (CommandResult, error)
}

// Command represents a command to be executed
type Command interface {
	// GetType returns the command type
	GetType() string
	
	// Validate validates the command parameters
	Validate() error
}

// CommandResult represents the result of a command execution
type CommandResult interface {
	// IsSuccess returns true if the command was successful
	IsSuccess() bool
	
	// GetMessage returns a human-readable message about the result
	GetMessage() string
	
	// GetData returns any data associated with the result
	GetData() interface{}
	
	// GetError returns any error that occurred
	GetError() error
}

// GenerateSchemaCommand represents a command to generate a schema
type GenerateSchemaCommand struct {
	FolderPath string
	OutputPath string
	Options    GenerateOptions
}

// UpdateSchemaCommand represents a command to update a schema
type UpdateSchemaCommand struct {
	FolderPath string
	SchemaPath string
	Options    UpdateOptions
}

// GenerateDataCommand represents a command to generate data
type GenerateDataCommand struct {
	FolderPath string
	SchemaPath string
	OutputPath string
	Options    DataGenerationOptions
}

// GenerateOptions defines options for schema generation
type GenerateOptions struct {
	IncludeMetadata    bool
	AutoDetectTypes    bool
	SkipEmptySheets    bool
	CustomHeaderRow    int
	ValidationRules    []models.ValidationRule
}

// UpdateOptions defines options for schema updates
type UpdateOptions struct {
	PreserveCustomRules bool
	UpdateTimestamps    bool
	MergeStrategy      string
	BackupExisting     bool
}

// DataGenerationOptions defines options for data generation
type DataGenerationOptions struct {
	IncludeMetadata     bool
	AutoGenerateIds     bool
	SkipValidation      bool
	CustomIdField       string
	OutputFormat        string
}

// SchemaCommandHandler handles schema-related commands
type SchemaCommandHandler interface {
	CommandHandler
	
	// HandleGenerate handles schema generation commands
	HandleGenerate(ctx context.Context, cmd *GenerateSchemaCommand) (*SchemaCommandResult, error)
	
	// HandleUpdate handles schema update commands
	HandleUpdate(ctx context.Context, cmd *UpdateSchemaCommand) (*SchemaCommandResult, error)
}

// DataCommandHandler handles data-related commands
type DataCommandHandler interface {
	CommandHandler
	
	// HandleGenerate handles data generation commands
	HandleGenerate(ctx context.Context, cmd *GenerateDataCommand) (*DataCommandResult, error)
}

// SchemaCommandResult represents the result of a schema command
type SchemaCommandResult struct {
	Success   bool
	Message   string
	Schema    *models.SchemaInfo
	FilePath  string
	Error     error
	Metadata  map[string]interface{}
}

// DataCommandResult represents the result of a data command
type DataCommandResult struct {
	Success    bool
	Message    string
	OutputData *models.OutputData
	FilePath   string
	Error      error
	Statistics *GenerationStatistics
}

// GenerationStatistics represents statistics about data generation
type GenerationStatistics struct {
	ProcessedFiles  int
	ProcessedSheets int
	GeneratedRecords int
	ProcessingTime   int64
	Errors          []string
}

// EventHandler defines the interface for handling application events
type EventHandler interface {
	// Handle processes an event
	Handle(ctx context.Context, event Event) error
}

// Event represents an application event
type Event interface {
	// GetType returns the event type
	GetType() string
	
	// GetTimestamp returns when the event occurred
	GetTimestamp() int64
	
	// GetData returns event data
	GetData() interface{}
}

// FileProcessedEvent represents an event when a file is processed
type FileProcessedEvent struct {
	Type      string
	Timestamp int64
	FilePath  string
	Success   bool
	Error     error
}

// SchemaUpdatedEvent represents an event when a schema is updated
type SchemaUpdatedEvent struct {
	Type      string
	Timestamp int64
	SchemaPath string
	Changes   []string
}

// DataGeneratedEvent represents an event when data is generated
type DataGeneratedEvent struct {
	Type       string
	Timestamp  int64
	OutputPath string
	Records    int
}

// ProgressHandler defines the interface for handling progress updates
type ProgressHandler interface {
	// Start starts progress tracking
	Start(ctx context.Context, total int, message string)
	
	// Update updates progress
	Update(ctx context.Context, current int, message string)
	
	// Complete completes progress tracking
	Complete(ctx context.Context, message string)
	
	// Error reports an error during progress
	Error(ctx context.Context, err error)
}

// ErrorHandler defines the interface for handling errors
type ErrorHandler interface {
	// Handle handles an error
	Handle(ctx context.Context, err error) error
	
	// ShouldRetry determines if an operation should be retried
	ShouldRetry(ctx context.Context, err error) bool
	
	// GetRetryDelay returns the delay before retrying
	GetRetryDelay(ctx context.Context, attempt int) int64
}

// Implementation methods for command interfaces

func (cmd *GenerateSchemaCommand) GetType() string {
	return "generate_schema"
}

func (cmd *GenerateSchemaCommand) Validate() error {
	if cmd.FolderPath == "" {
		return NewValidationError("folder path is required")
	}
	return nil
}

func (cmd *UpdateSchemaCommand) GetType() string {
	return "update_schema"
}

func (cmd *UpdateSchemaCommand) Validate() error {
	if cmd.FolderPath == "" {
		return NewValidationError("folder path is required")
	}
	if cmd.SchemaPath == "" {
		return NewValidationError("schema path is required")
	}
	return nil
}

func (cmd *GenerateDataCommand) GetType() string {
	return "generate_data"
}

func (cmd *GenerateDataCommand) Validate() error {
	if cmd.FolderPath == "" {
		return NewValidationError("folder path is required")
	}
	if cmd.SchemaPath == "" {
		return NewValidationError("schema path is required")
	}
	return nil
}

// Implementation methods for result interfaces

func (r *SchemaCommandResult) IsSuccess() bool {
	return r.Success
}

func (r *SchemaCommandResult) GetMessage() string {
	return r.Message
}

func (r *SchemaCommandResult) GetData() interface{} {
	return r.Schema
}

func (r *SchemaCommandResult) GetError() error {
	return r.Error
}

func (r *DataCommandResult) IsSuccess() bool {
	return r.Success
}

func (r *DataCommandResult) GetMessage() string {
	return r.Message
}

func (r *DataCommandResult) GetData() interface{} {
	return r.OutputData
}

func (r *DataCommandResult) GetError() error {
	return r.Error
}

// Implementation methods for event interfaces

func (e *FileProcessedEvent) GetType() string {
	return e.Type
}

func (e *FileProcessedEvent) GetTimestamp() int64 {
	return e.Timestamp
}

func (e *FileProcessedEvent) GetData() interface{} {
	return map[string]interface{}{
		"file_path": e.FilePath,
		"success":   e.Success,
		"error":     e.Error,
	}
}

func (e *SchemaUpdatedEvent) GetType() string {
	return e.Type
}

func (e *SchemaUpdatedEvent) GetTimestamp() int64 {
	return e.Timestamp
}

func (e *SchemaUpdatedEvent) GetData() interface{} {
	return map[string]interface{}{
		"schema_path": e.SchemaPath,
		"changes":     e.Changes,
	}
}

func (e *DataGeneratedEvent) GetType() string {
	return e.Type
}

func (e *DataGeneratedEvent) GetTimestamp() int64 {
	return e.Timestamp
}

func (e *DataGeneratedEvent) GetData() interface{} {
	return map[string]interface{}{
		"output_path": e.OutputPath,
		"records":     e.Records,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}