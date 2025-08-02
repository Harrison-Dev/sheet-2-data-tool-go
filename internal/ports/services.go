package ports

import (
	"context"

	"excel-schema-generator/internal/core/models"
)

// SchemaService defines the interface for schema-related business operations
type SchemaService interface {
	// GenerateFromFolder generates a new schema from Excel files in a folder
	GenerateFromFolder(ctx context.Context, folderPath string) (*models.SchemaInfo, error)
	
	// UpdateFromFolder updates an existing schema with Excel files from a folder
	UpdateFromFolder(ctx context.Context, schema *models.SchemaInfo, folderPath string) error
	
	// Validate validates a schema for consistency and completeness
	Validate(ctx context.Context, schema *models.SchemaInfo) error
	
	// Merge merges two schemas together
	Merge(ctx context.Context, base, additional *models.SchemaInfo) (*models.SchemaInfo, error)
	
	// GetSchemaStatistics returns statistics about a schema
	GetSchemaStatistics(ctx context.Context, schema *models.SchemaInfo) (*SchemaStatistics, error)
}

// DataService defines the interface for data processing operations
type DataService interface {
	// GenerateFromSchema generates JSON data from Excel files using a schema
	GenerateFromSchema(ctx context.Context, schema *models.SchemaInfo, folderPath string) (*models.OutputData, error)
	
	// ExtractFromFile extracts data from a single Excel file
	ExtractFromFile(ctx context.Context, filePath string, fileInfo models.ExcelFileInfo) (map[string][]interface{}, error)
	
	// Transform transforms raw Excel data according to schema rules
	Transform(ctx context.Context, rawData *models.ExcelData, sheetInfo models.SheetInfo) ([]interface{}, error)
	
	// ValidateData validates extracted data against schema rules
	ValidateData(ctx context.Context, data []interface{}, sheetInfo models.SheetInfo) error
}

// ExcelService defines the interface for Excel file processing operations
type ExcelService interface {
	// ProcessFile processes a single Excel file and extracts its structure
	ProcessFile(ctx context.Context, filePath string) (*models.ExcelData, error)
	
	// ProcessFolder processes all Excel files in a folder
	ProcessFolder(ctx context.Context, folderPath string) (map[string]*models.ExcelData, error)
	
	// GetFileChecksum calculates checksum for an Excel file
	GetFileChecksum(ctx context.Context, filePath string) (string, error)
	
	// DetectChanges detects changes in Excel files compared to schema
	DetectChanges(ctx context.Context, schema *models.SchemaInfo, folderPath string) (*ChangeReport, error)
}

// ValidationService defines the interface for validation operations
type ValidationService interface {
	// ValidateExcelFile validates an Excel file structure
	ValidateExcelFile(ctx context.Context, filePath string) error
	
	// ValidateSchema validates a schema structure
	ValidateSchema(ctx context.Context, schema *models.SchemaInfo) error
	
	// ValidateDataTypes validates data types in extracted data
	ValidateDataTypes(ctx context.Context, data []interface{}, fields []models.DataClassInfo) error
	
	// ValidateRules validates custom validation rules
	ValidateRules(ctx context.Context, data []interface{}, rules []models.ValidationRule) error
}

// ConfigService defines the interface for configuration management
type ConfigService interface {
	// Load loads configuration from default location
	Load(ctx context.Context) (*AppConfig, error)
	
	// Save saves configuration to default location
	Save(ctx context.Context, config *AppConfig) error
	
	// GetDefaults returns default configuration values
	GetDefaults() *AppConfig
	
	// Validate validates configuration values
	Validate(ctx context.Context, config *AppConfig) error
}

// LoggingService defines the interface for logging operations
type LoggingService interface {
	// Debug logs a debug message
	Debug(msg string, keysAndValues ...any)
	
	// Info logs an info message
	Info(msg string, keysAndValues ...any)
	
	// Warn logs a warning message
	Warn(msg string, keysAndValues ...any)
	
	// Error logs an error message
	Error(msg string, keysAndValues ...any)
	
	// With returns a new logger with additional context
	With(keysAndValues ...any) LoggingService
}

// SchemaStatistics represents statistics about a schema
type SchemaStatistics struct {
	FileCount      int
	SheetCount     int
	FieldCount     int
	TotalRows      int
	LastUpdated    int64
	ValidationErrors []string
}

// ChangeReport represents changes detected in Excel files
type ChangeReport struct {
	AddedFiles    []string
	ModifiedFiles []string
	RemovedFiles  []string
	AddedSheets   map[string][]string
	ModifiedSheets map[string][]string
	RemovedSheets map[string][]string
}

// AppConfig represents application configuration
type AppConfig struct {
	ExcelFolder  string `json:"excel_folder"`
	SchemaFolder string `json:"schema_folder"`
	OutputFolder string `json:"output_folder"`
	LogLevel     string `json:"log_level"`
	LogFormat    string `json:"log_format"`
}

// ServiceManager defines the interface for managing all services
type ServiceManager interface {
	// Schema returns the schema service
	Schema() SchemaService
	
	// Data returns the data service
	Data() DataService
	
	// Excel returns the Excel service
	Excel() ExcelService
	
	// Validation returns the validation service
	Validation() ValidationService
	
	// Config returns the config service
	Config() ConfigService
	
	// Logging returns the logging service
	Logging() LoggingService
}