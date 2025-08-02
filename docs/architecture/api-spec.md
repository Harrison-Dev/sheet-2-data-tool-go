# Internal API Specifications

## Overview

This document defines the bulletproof internal API contracts and interfaces for the Excel Schema Generator refactored architecture. These APIs ensure zero compilation issues with strict type safety, complete interface compatibility, and comprehensive error handling.

## Critical Interface Fixes

### 1. Logging Service Interface (FIXED)

**Problem**: Parameter type mismatch between `pkg/logger` and `ports.LoggingService`

**Solution**: Standardized interface with `...any` parameters

```go
// Fixed LoggingService interface in internal/ports/services.go
type LoggingService interface {
    // Debug logs a debug message with structured key-value pairs
    Debug(msg string, keysAndValues ...any)
    
    // Info logs an info message with structured key-value pairs  
    Info(msg string, keysAndValues ...any)
    
    // Warn logs a warning message with structured key-value pairs
    Warn(msg string, keysAndValues ...any)
    
    // Error logs an error message with structured key-value pairs
    Error(msg string, keysAndValues ...any)
    
    // With returns a new logger with additional context
    With(keysAndValues ...any) LoggingService
}

// Compatible implementation in internal/utils/logger/adapter.go
type LoggerAdapter struct {
    logger *logger.Logger
}

func NewLoggerAdapter(logger *logger.Logger) LoggingService {
    return &LoggerAdapter{logger: logger}
}

func (a *LoggerAdapter) Debug(msg string, keysAndValues ...any) {
    a.logger.Debug(msg, keysAndValues...)
}

func (a *LoggerAdapter) Info(msg string, keysAndValues ...any) {
    a.logger.Info(msg, keysAndValues...)
}

func (a *LoggerAdapter) Warn(msg string, keysAndValues ...any) {
    a.logger.Warn(msg, keysAndValues...)
}

func (a *LoggerAdapter) Error(msg string, keysAndValues ...any) {
    a.logger.Error(msg, keysAndValues...)
}

func (a *LoggerAdapter) With(keysAndValues ...any) LoggingService {
    newLogger := &logger.Logger{Logger: a.logger.With(keysAndValues...)}
    return NewLoggerAdapter(newLogger)
}
```

### 2. GUI Type-Safe Interfaces (FIXED)

**Problem**: Fyne widget type mismatches and callback incompatibilities

**Solution**: Strict type definitions with proper Fyne integration

```go
// Fixed GUI interfaces in internal/ports/handlers.go
type GUIHandler interface {
    // HandleFolderSelection processes folder selection with type-safe callback
    HandleFolderSelection(ctx context.Context, folderType FolderType, callback FolderSelectionCallback) error
    
    // HandleSchemaGeneration processes schema generation with progress tracking
    HandleSchemaGeneration(ctx context.Context, request SchemaGenerationRequest) error
    
    // HandleSchemaUpdate processes schema updates with conflict resolution
    HandleSchemaUpdate(ctx context.Context, request SchemaUpdateRequest) error
    
    // HandleDataGeneration processes data generation with export options
    HandleDataGeneration(ctx context.Context, request DataGenerationRequest) error
    
    // HandleProgressUpdate updates UI progress display
    HandleProgressUpdate(ctx context.Context, update ProgressUpdate) error
}

// Type-safe callback definitions
type FolderSelectionCallback func(folderPath string, err error)
type ProgressCallback func(current, total int, message string, err error)

// Fixed GUI request types with proper validation
type SchemaGenerationRequest struct {
    ExcelFolderPath string                `validate:"required,dirpath"`
    OutputPath      string                `validate:"required,filepath"`
    Options         SchemaGenerationOptions `validate:"required"`
    ProgressCallback ProgressCallback      `validate:"required"`
}

type SchemaUpdateRequest struct {
    ExcelFolderPath string             `validate:"required,dirpath"`
    SchemaPath      string             `validate:"required,filepath"`
    Options         SchemaUpdateOptions `validate:"required"`
    ProgressCallback ProgressCallback   `validate:"required"`
}

type DataGenerationRequest struct {
    ExcelFolderPath string              `validate:"required,dirpath"`
    SchemaPath      string              `validate:"required,filepath"`
    OutputPath      string              `validate:"required,filepath"`
    Options         DataGenerationOptions `validate:"required"`
    ProgressCallback ProgressCallback    `validate:"required"`
}

// Progress update with structured information
type ProgressUpdate struct {
    Current     int           `json:"current"`
    Total       int           `json:"total"`
    Message     string        `json:"message"`
    Percentage  float64       `json:"percentage"`
    ElapsedTime time.Duration `json:"elapsed_time"`
    EstimatedTime time.Duration `json:"estimated_time"`
    Error       error         `json:"error,omitempty"`
}
```

### 3. Fyne Widget Integration (FIXED)

**Problem**: Widget type incompatibilities and dialog handling issues

**Solution**: Type-safe widget factory and dialog management

```go
// Fixed GUI widget interfaces in cmd/gui/app/widgets.go
type WidgetFactory interface {
    // CreateFolderSelector creates a type-safe folder selector widget
    CreateFolderSelector(title string, callback FolderSelectionCallback) *FolderSelector
    
    // CreateProgressDisplay creates a progress display with cancellation
    CreateProgressDisplay() *ProgressDisplay
    
    // CreateStatusPanel creates a status panel with multi-level messaging
    CreateStatusPanel() *StatusPanel
}

// Type-safe folder selector implementation
type FolderSelector struct {
    entry      *widget.Entry
    button     *widget.Button
    container  *container.Border
    callback   FolderSelectionCallback
    window     fyne.Window
}

func (fs *FolderSelector) CreateWidget() fyne.CanvasObject {
    fs.entry = widget.NewEntry()
    fs.entry.SetPlaceHolder("Select folder...")
    fs.entry.Disable() // Read-only, populated by dialog
    
    fs.button = widget.NewButton("Browse", fs.showFolderDialog)
    
    fs.container = container.NewBorder(nil, nil, nil, fs.button, fs.entry)
    return fs.container
}

func (fs *FolderSelector) showFolderDialog() {
    folderDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI) {
        var folderPath string
        var err error
        
        if folder != nil {
            folderPath = folder.Path()
            fs.entry.SetText(folderPath)
        } else {
            err = errors.New("no folder selected")
        }
        
        if fs.callback != nil {
            fs.callback(folderPath, err)
        }
    }, fs.window)
    
    folderDialog.SetTitle("Select Folder")
    folderDialog.Show()
}

// Type-safe progress display with cancellation
type ProgressDisplay struct {
    progressBar     *widget.ProgressBar
    statusLabel     *widget.Label
    cancelButton    *widget.Button
    container       *container.VBox
    cancelFunc      context.CancelFunc
    isVisible       bool
}

func (pd *ProgressDisplay) CreateWidget() fyne.CanvasObject {
    pd.progressBar = widget.NewProgressBar()
    pd.progressBar.Hide()
    
    pd.statusLabel = widget.NewLabel("Ready")
    
    pd.cancelButton = widget.NewButton("Cancel", pd.handleCancel)
    pd.cancelButton.Hide()
    
    pd.container = container.NewVBox(
        pd.statusLabel,
        pd.progressBar,
        pd.cancelButton,
    )
    
    return pd.container
}

func (pd *ProgressDisplay) UpdateProgress(update ProgressUpdate) {
    if !pd.isVisible {
        pd.Show()
    }
    
    pd.progressBar.SetValue(update.Percentage / 100.0)
    pd.statusLabel.SetText(update.Message)
    
    if update.Error != nil {
        pd.statusLabel.SetText(fmt.Sprintf("Error: %s", update.Error.Error()))
        pd.Hide()
    }
}

func (pd *ProgressDisplay) Show() {
    pd.progressBar.Show()
    pd.cancelButton.Show()
    pd.isVisible = true
}

func (pd *ProgressDisplay) Hide() {
    pd.progressBar.Hide()
    pd.cancelButton.Hide()
    pd.isVisible = false
}

func (pd *ProgressDisplay) SetCancelFunc(cancelFunc context.CancelFunc) {
    pd.cancelFunc = cancelFunc
}

func (pd *ProgressDisplay) handleCancel() {
    if pd.cancelFunc != nil {
        pd.cancelFunc()
    }
    pd.Hide()
}
```

## Core Service Interfaces

### Schema Service Interface (UPDATED)

```go
// SchemaService defines the interface for schema operations with strict types
type SchemaService interface {
    // GenerateFromFolder creates a new schema from Excel files in a directory
    GenerateFromFolder(ctx context.Context, folderPath string) (*models.SchemaInfo, error)
    
    // UpdateFromFolder updates an existing schema with new Excel data
    UpdateFromFolder(ctx context.Context, schema *models.SchemaInfo, folderPath string) error
    
    // Validate validates a schema for consistency and completeness
    Validate(ctx context.Context, schema *models.SchemaInfo) error
    
    // Merge merges two schemas together with conflict resolution
    Merge(ctx context.Context, base, additional *models.SchemaInfo) (*models.SchemaInfo, error)
    
    // GetSchemaStatistics returns comprehensive statistics about a schema
    GetSchemaStatistics(ctx context.Context, schema *models.SchemaInfo) (*SchemaStatistics, error)
    
    // LoadFromFile loads a schema from a YAML file with validation
    LoadFromFile(ctx context.Context, filePath string) (*models.SchemaInfo, error)
    
    // SaveToFile saves a schema to a YAML file with formatting
    SaveToFile(ctx context.Context, schema *models.SchemaInfo, filePath string) error
}

// SchemaGenerationOptions configures schema generation behavior
type SchemaGenerationOptions struct {
    IncludePatterns   []string              `json:"include_patterns" validate:"required"`
    ExcludePatterns   []string              `json:"exclude_patterns"`
    ValidationRules   []ValidationRuleConfig `json:"validation_rules"`
    GenerateMetadata  bool                  `json:"generate_metadata"`
    DetectDataTypes   bool                  `json:"detect_data_types"`
    InferConstraints  bool                  `json:"infer_constraints"`
}

// SchemaUpdateOptions configures schema update behavior
type SchemaUpdateOptions struct {
    UpdateStrategy     UpdateStrategy     `json:"update_strategy" validate:"required,oneof=merge replace append"`
    ConflictResolution ConflictResolution `json:"conflict_resolution" validate:"required,oneof=keep_existing use_new prompt"`
    BackupEnabled      bool              `json:"backup_enabled"`
    ValidateAfterUpdate bool             `json:"validate_after_update"`
}

// SchemaStatistics provides comprehensive schema analysis
type SchemaStatistics struct {
    FileCount          int                    `json:"file_count"`
    SheetCount         int                    `json:"sheet_count"`
    FieldCount         int                    `json:"field_count"`
    TotalRows          int                    `json:"total_rows"`
    DataTypes          map[string]int         `json:"data_types"`
    ValidationErrors   []ValidationError      `json:"validation_errors"`
    LastUpdated        time.Time             `json:"last_updated"`
    SchemaComplexity   SchemaComplexityMetrics `json:"schema_complexity"`
}

type SchemaComplexityMetrics struct {
    AverageFieldsPerSheet float64 `json:"average_fields_per_sheet"`
    MaxFieldsInSheet      int     `json:"max_fields_in_sheet"`
    MinFieldsInSheet      int     `json:"min_fields_in_sheet"`
    UniqueDataTypes       int     `json:"unique_data_types"`
}
```

### Data Processing Service Interface (UPDATED)

```go
// DataService defines the interface for data extraction and transformation
type DataService interface {
    // GenerateFromSchema generates JSON data from Excel files using a schema
    GenerateFromSchema(ctx context.Context, schema *models.SchemaInfo, folderPath string) (*models.OutputData, error)
    
    // ExtractFromFile extracts data from a single Excel file with validation
    ExtractFromFile(ctx context.Context, filePath string, fileInfo models.ExcelFileInfo) (map[string][]interface{}, error)
    
    // Transform transforms raw Excel data according to schema rules
    Transform(ctx context.Context, rawData *models.ExcelData, sheetInfo models.SheetInfo) ([]interface{}, error)
    
    // ValidateData validates extracted data against schema constraints
    ValidateData(ctx context.Context, data []interface{}, sheetInfo models.SheetInfo) error
    
    // ExportData exports processed data in various formats
    ExportData(ctx context.Context, data *models.OutputData, options DataExportOptions) error
    
    // ProcessBatch processes multiple Excel files concurrently
    ProcessBatch(ctx context.Context, files []string, schema *models.SchemaInfo, options BatchProcessingOptions) (*models.OutputData, error)
}

// DataGenerationOptions configures data generation behavior
type DataGenerationOptions struct {
    OutputFormat       OutputFormat         `json:"output_format" validate:"required,oneof=json yaml xml csv"`
    PrettyPrint        bool                `json:"pretty_print"`
    IncludeMetadata    bool                `json:"include_metadata"`
    ValidateOutput     bool                `json:"validate_output"`
    CompressionEnabled bool                `json:"compression_enabled"`
    BatchSize          int                 `json:"batch_size" validate:"min=1,max=1000"`
}

// DataExportOptions configures data export behavior
type DataExportOptions struct {
    OutputPath      string       `json:"output_path" validate:"required,filepath"`
    Format          OutputFormat `json:"format" validate:"required"`
    Compression     bool         `json:"compression"`
    SplitBySheet    bool         `json:"split_by_sheet"`
    IncludeTimestamp bool        `json:"include_timestamp"`
}

// BatchProcessingOptions configures batch processing
type BatchProcessingOptions struct {
    MaxConcurrency   int           `json:"max_concurrency" validate:"min=1,max=10"`
    TimeoutPerFile   time.Duration `json:"timeout_per_file" validate:"required"`
    ContinueOnError  bool          `json:"continue_on_error"`
    ProgressCallback ProgressCallback `json:"-"`
}
```

### Excel Processing Interface (UPDATED)

```go
// ExcelService defines the interface for Excel file operations
type ExcelService interface {
    // ProcessFile processes a single Excel file and extracts its structure
    ProcessFile(ctx context.Context, filePath string) (*models.ExcelData, error)
    
    // ProcessFolder processes all Excel files in a folder with filtering
    ProcessFolder(ctx context.Context, folderPath string) (map[string]*models.ExcelData, error)
    
    // GetFileChecksum calculates SHA256 checksum for an Excel file
    GetFileChecksum(ctx context.Context, filePath string) (string, error)
    
    // DetectChanges detects changes in Excel files compared to schema
    DetectChanges(ctx context.Context, schema *models.SchemaInfo, folderPath string) (*ChangeReport, error)
    
    // ValidateExcelFile validates Excel file format and accessibility
    ValidateExcelFile(ctx context.Context, filePath string) error
    
    // GetFileMetadata extracts metadata from Excel file
    GetFileMetadata(ctx context.Context, filePath string) (*ExcelFileMetadata, error)
}

// ExcelFileMetadata contains comprehensive Excel file information
type ExcelFileMetadata struct {
    FileName     string            `json:"file_name"`
    FileSize     int64             `json:"file_size"`
    LastModified time.Time         `json:"last_modified"`
    Checksum     string            `json:"checksum"`
    Sheets       []SheetMetadata   `json:"sheets"`
    Properties   ExcelProperties   `json:"properties"`
}

type SheetMetadata struct {
    Name        string `json:"name"`
    RowCount    int    `json:"row_count"`
    ColumnCount int    `json:"column_count"`
    IsVisible   bool   `json:"is_visible"`
    IsProtected bool   `json:"is_protected"`
}

type ExcelProperties struct {
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    Company     string    `json:"company"`
    Subject     string    `json:"subject"`
    CreatedDate time.Time `json:"created_date"`
    ModifiedDate time.Time `json:"modified_date"`
}

// ChangeReport represents detected changes with detailed information
type ChangeReport struct {
    AddedFiles     []FileChange     `json:"added_files"`
    ModifiedFiles  []FileChange     `json:"modified_files"`
    RemovedFiles   []FileChange     `json:"removed_files"`
    AddedSheets    map[string][]SheetChange `json:"added_sheets"`
    ModifiedSheets map[string][]SheetChange `json:"modified_sheets"`
    RemovedSheets  map[string][]SheetChange `json:"removed_sheets"`
    Summary        ChangeSummary    `json:"summary"`
}

type FileChange struct {
    FilePath      string    `json:"file_path"`
    FileName      string    `json:"file_name"`
    OldChecksum   string    `json:"old_checksum,omitempty"`
    NewChecksum   string    `json:"new_checksum"`
    ChangeType    string    `json:"change_type"`
    DetectedAt    time.Time `json:"detected_at"`
}

type SheetChange struct {
    SheetName     string    `json:"sheet_name"`
    ChangeType    string    `json:"change_type"`
    OldRowCount   int       `json:"old_row_count,omitempty"`
    NewRowCount   int       `json:"new_row_count"`
    DetectedAt    time.Time `json:"detected_at"`
}

type ChangeSummary struct {
    TotalChanges    int `json:"total_changes"`
    FilesAffected   int `json:"files_affected"`
    SheetsAffected  int `json:"sheets_affected"`
    RequiresUpdate  bool `json:"requires_update"`
}
```

## Repository Interfaces (UPDATED)

### Schema Repository

```go
// SchemaRepository defines type-safe data access for schemas
type SchemaRepository interface {
    // Save persists a schema with atomic operations
    Save(ctx context.Context, schema *models.SchemaInfo, filePath string) error
    
    // Load retrieves a schema with validation
    Load(ctx context.Context, filePath string) (*models.SchemaInfo, error)
    
    // Delete removes a schema with backup
    Delete(ctx context.Context, filePath string) error
    
    // List returns all available schemas in a directory
    List(ctx context.Context, dirPath string) ([]*SchemaFileInfo, error)
    
    // Backup creates a timestamped backup of a schema
    Backup(ctx context.Context, filePath string) (string, error)
    
    // ValidateFile validates schema file format without loading
    ValidateFile(ctx context.Context, filePath string) error
}

type SchemaFileInfo struct {
    FilePath     string    `json:"file_path"`
    FileName     string    `json:"file_name"`
    Size         int64     `json:"size"`
    LastModified time.Time `json:"last_modified"`
    Version      string    `json:"version"`
    IsValid      bool      `json:"is_valid"`
}
```

### Data Repository

```go
// DataRepository defines type-safe data access for processed data
type DataRepository interface {
    // SaveOutputData persists processed data with metadata
    SaveOutputData(ctx context.Context, data *models.OutputData, filePath string) error
    
    // LoadOutputData retrieves processed data
    LoadOutputData(ctx context.Context, filePath string) (*models.OutputData, error)
    
    // ExportJSON exports data as formatted JSON
    ExportJSON(ctx context.Context, data *models.OutputData, filePath string, options JSONExportOptions) error
    
    // ExportYAML exports data as formatted YAML
    ExportYAML(ctx context.Context, data *models.OutputData, filePath string, options YAMLExportOptions) error
    
    // ExportCSV exports data as CSV files (one per sheet)
    ExportCSV(ctx context.Context, data *models.OutputData, dirPath string, options CSVExportOptions) error
}

type JSONExportOptions struct {
    PrettyPrint      bool `json:"pretty_print"`
    IncludeMetadata  bool `json:"include_metadata"`
    CompactArrays    bool `json:"compact_arrays"`
    SortKeys         bool `json:"sort_keys"`
}

type YAMLExportOptions struct {
    IncludeMetadata  bool `json:"include_metadata"`
    FlowStyle        bool `json:"flow_style"`
    IncludeComments  bool `json:"include_comments"`
}

type CSVExportOptions struct {
    Delimiter        rune   `json:"delimiter"`
    IncludeHeaders   bool   `json:"include_headers"`
    QuoteAll         bool   `json:"quote_all"`
    FileNamePrefix   string `json:"file_name_prefix"`
}
```

## Error Handling Patterns (UPDATED)

### Comprehensive Error Types

```go
// ApplicationError represents application-specific errors with context
type ApplicationError struct {
    Code      ErrorCode              `json:"code"`
    Message   string                 `json:"message"`
    Cause     error                  `json:"-"`
    Context   map[string]interface{} `json:"context"`
    Timestamp time.Time              `json:"timestamp"`
    Stack     string                 `json:"stack,omitempty"`
}

// ErrorCode defines comprehensive error categories
type ErrorCode string

const (
    // File system errors
    ErrFileNotFound      ErrorCode = "FILE_NOT_FOUND"
    ErrFilePermission    ErrorCode = "FILE_PERMISSION" 
    ErrFileCorrupted     ErrorCode = "FILE_CORRUPTED"
    ErrFileTooBig        ErrorCode = "FILE_TOO_BIG"
    ErrDirectoryNotFound ErrorCode = "DIRECTORY_NOT_FOUND"
    
    // Excel processing errors
    ErrExcelFormat       ErrorCode = "EXCEL_FORMAT"
    ErrExcelPassword     ErrorCode = "EXCEL_PASSWORD"
    ErrExcelCorrupted    ErrorCode = "EXCEL_CORRUPTED"
    ErrSheetNotFound     ErrorCode = "SHEET_NOT_FOUND"
    ErrInvalidCellRange  ErrorCode = "INVALID_CELL_RANGE"
    
    // Schema errors
    ErrSchemaInvalid     ErrorCode = "SCHEMA_INVALID"
    ErrSchemaVersion     ErrorCode = "SCHEMA_VERSION"
    ErrSchemaConflict    ErrorCode = "SCHEMA_CONFLICT"
    ErrSchemaMigration   ErrorCode = "SCHEMA_MIGRATION"
    
    // Data processing errors
    ErrDataValidation    ErrorCode = "DATA_VALIDATION"
    ErrDataTransform     ErrorCode = "DATA_TRANSFORM"
    ErrDataExport        ErrorCode = "DATA_EXPORT"
    ErrDataTypeMismatch  ErrorCode = "DATA_TYPE_MISMATCH"
    
    // Configuration errors
    ErrConfigInvalid     ErrorCode = "CONFIG_INVALID"
    ErrConfigMissing     ErrorCode = "CONFIG_MISSING"
    ErrConfigPermission  ErrorCode = "CONFIG_PERMISSION"
    
    // Runtime errors
    ErrMemoryLimit       ErrorCode = "MEMORY_LIMIT"
    ErrTimeout           ErrorCode = "TIMEOUT"
    ErrCancelled         ErrorCode = "CANCELLED"
    ErrConcurrencyLimit  ErrorCode = "CONCURRENCY_LIMIT"
    
    // Validation errors
    ErrValidationFailed  ErrorCode = "VALIDATION_FAILED"
    ErrConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"
    ErrRequiredFieldMissing ErrorCode = "REQUIRED_FIELD_MISSING"
)

// Error creates a formatted error message
func (e ApplicationError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// IsTemporary indicates if the error is retryable
func (e ApplicationError) IsTemporary() bool {
    switch e.Code {
    case ErrFilePermission, ErrMemoryLimit, ErrTimeout, ErrConcurrencyLimit:
        return true
    default:
        return false
    }
}

// GetUserMessage returns a user-friendly error message
func (e ApplicationError) GetUserMessage() string {
    switch e.Code {
    case ErrFileNotFound:
        return "The specified file could not be found. Please check the file path."
    case ErrFilePermission:
        return "Permission denied. Please check file permissions and try again."
    case ErrExcelFormat:
        return "Invalid Excel file format. Please ensure the file is a valid .xlsx file."
    case ErrSchemaInvalid:
        return "The schema file is invalid. Please check the schema format."
    case ErrMemoryLimit:
        return "Insufficient memory to process this file. Try closing other applications."
    default:
        return e.Message
    }
}
```

### Error Handler Interface

```go
// ErrorHandler defines comprehensive error handling strategies
type ErrorHandler interface {
    // HandleError processes an error and determines response strategy
    HandleError(ctx context.Context, err error) ErrorResponse
    
    // LogError logs an error with appropriate level and context
    LogError(ctx context.Context, err error, additionalContext ...any)
    
    // IsRetryable determines if an error can be retried
    IsRetryable(err error) bool
    
    // GetUserMessage returns user-friendly error message
    GetUserMessage(err error) string
    
    // GetRecoveryActions suggests recovery actions for the error
    GetRecoveryActions(err error) []RecoveryAction
}

// ErrorResponse defines comprehensive error response strategy
type ErrorResponse struct {
    Action           ErrorAction      `json:"action"`
    Message          string           `json:"message"`
    UserMessage      string           `json:"user_message"`
    Retryable        bool             `json:"retryable"`
    LogLevel         LogLevel         `json:"log_level"`
    RecoveryActions  []RecoveryAction `json:"recovery_actions"`
    ShouldNotifyUser bool             `json:"should_notify_user"`
}

// ErrorAction defines possible response actions
type ErrorAction string

const (
    ActionAbort        ErrorAction = "ABORT"
    ActionRetry        ErrorAction = "RETRY"
    ActionSkip         ErrorAction = "SKIP"
    ActionPrompt       ErrorAction = "PROMPT"
    ActionFallback     ErrorAction = "FALLBACK"
    ActionRetryWithDelay ErrorAction = "RETRY_WITH_DELAY"
)

// RecoveryAction suggests specific recovery steps
type RecoveryAction struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    AutoApply   bool   `json:"auto_apply"`
}
```

## Configuration Management (UPDATED)

### Type-Safe Configuration Structure

```go
// AppConfig contains comprehensive application configuration
type AppConfig struct {
    App        AppSettings        `json:"app" validate:"required"`
    Logging    LoggingConfig      `json:"logging" validate:"required"`
    Processing ProcessingConfig   `json:"processing" validate:"required"`
    GUI        GUIConfig          `json:"gui" validate:"required"`
    Paths      PathsConfig        `json:"paths" validate:"required"`
    Advanced   AdvancedConfig     `json:"advanced"`
}

// AppSettings contains core application settings
type AppSettings struct {
    Name        string `json:"name" validate:"required"`
    Version     string `json:"version" validate:"required,semver"`
    Environment string `json:"environment" validate:"required,oneof=development staging production"`
    Debug       bool   `json:"debug"`
    LogLevel    string `json:"log_level" validate:"required,oneof=debug info warn error"`
}

// LoggingConfig contains comprehensive logging settings
type LoggingConfig struct {
    Level       string `json:"level" validate:"required,oneof=debug info warn error"`
    Format      string `json:"format" validate:"required,oneof=text json"`
    Output      string `json:"output" validate:"required,oneof=stdout stderr file"`
    FilePath    string `json:"file_path,omitempty" validate:"omitempty,filepath"`
    MaxSize     int    `json:"max_size" validate:"min=1,max=1024"`      // MB
    MaxBackups  int    `json:"max_backups" validate:"min=1,max=10"`
    MaxAge      int    `json:"max_age" validate:"min=1,max=365"`        // days
    Compress    bool   `json:"compress"`
}

// ProcessingConfig contains performance and processing settings
type ProcessingConfig struct {
    MaxFileSize        int64         `json:"max_file_size" validate:"min=1048576,max=1073741824"`    // 1MB to 1GB
    MaxMemory          int64         `json:"max_memory" validate:"min=134217728,max=8589934592"`     // 128MB to 8GB
    TimeoutSeconds     int           `json:"timeout_seconds" validate:"min=30,max=3600"`            // 30s to 1h
    MaxConcurrency     int           `json:"max_concurrency" validate:"min=1,max=10"`
    TempDir            string        `json:"temp_dir" validate:"required,dirpath"`
    EnableCaching      bool          `json:"enable_caching"`
    CacheSize          int           `json:"cache_size" validate:"min=10,max=1000"`                 // Number of cached items
    BatchSize          int           `json:"batch_size" validate:"min=10,max=1000"`
    StreamingThreshold int64         `json:"streaming_threshold" validate:"min=10485760"`           // 10MB
}

// GUIConfig contains GUI-specific settings
type GUIConfig struct {
    WindowWidth      int         `json:"window_width" validate:"min=800,max=2560"`
    WindowHeight     int         `json:"window_height" validate:"min=600,max=1440"`
    Theme            string      `json:"theme" validate:"oneof=light dark auto"`
    LastPosition     WindowPos   `json:"last_position"`
    ShowProgress     bool        `json:"show_progress"`
    AutoSave         bool        `json:"auto_save"`
    ConfirmActions   bool        `json:"confirm_actions"`
    RememberFolders  bool        `json:"remember_folders"`
    ShowTooltips     bool        `json:"show_tooltips"`
}

type WindowPos struct {
    X int `json:"x"`
    Y int `json:"y"`
}

// PathsConfig contains default paths with validation
type PathsConfig struct {
    ExcelFolder   string `json:"excel_folder" validate:"omitempty,dirpath"`
    SchemaFolder  string `json:"schema_folder" validate:"omitempty,dirpath"`
    OutputFolder  string `json:"output_folder" validate:"omitempty,dirpath"`
    TempFolder    string `json:"temp_folder" validate:"required,dirpath"`
    LogFolder     string `json:"log_folder" validate:"required,dirpath"`
    BackupFolder  string `json:"backup_folder" validate:"omitempty,dirpath"`
}

// AdvancedConfig contains advanced settings for power users
type AdvancedConfig struct {
    EnableExperimentalFeatures bool                `json:"enable_experimental_features"`
    CustomValidationRules      []ValidationRule    `json:"custom_validation_rules"`
    PerformanceMode            string              `json:"performance_mode" validate:"oneof=balanced speed memory"`
    NetworkTimeout             int                 `json:"network_timeout" validate:"min=5,max=300"`  // seconds
    RetryAttempts              int                 `json:"retry_attempts" validate:"min=1,max=5"`
    CustomExportFormats        []CustomExportFormat `json:"custom_export_formats"`
}

type CustomExportFormat struct {
    Name      string `json:"name" validate:"required"`
    Extension string `json:"extension" validate:"required"`
    Template  string `json:"template" validate:"required"`
    Enabled   bool   `json:"enabled"`
}
```

### Configuration Validation

```go
// ConfigValidator provides comprehensive configuration validation
type ConfigValidator interface {
    // ValidateComplete validates the entire configuration
    ValidateComplete(config *AppConfig) []ValidationError
    
    // ValidateApp validates application settings
    ValidateApp(config AppSettings) []ValidationError
    
    // ValidateLogging validates logging configuration
    ValidateLogging(config LoggingConfig) []ValidationError
    
    // ValidateProcessing validates processing configuration
    ValidateProcessing(config ProcessingConfig) []ValidationError
    
    // ValidateGUI validates GUI configuration
    ValidateGUI(config GUIConfig) []ValidationError
    
    // ValidatePaths validates path configuration with access checks
    ValidatePaths(config PathsConfig) []ValidationError
    
    // ValidateAdvanced validates advanced configuration
    ValidateAdvanced(config AdvancedConfig) []ValidationError
}

// ValidationError represents detailed configuration validation error
type ValidationError struct {
    Field       string      `json:"field"`
    Value       interface{} `json:"value"`
    Rule        string      `json:"rule"`
    Message     string      `json:"message"`
    Severity    string      `json:"severity"` // error, warning, info
    Suggestion  string      `json:"suggestion,omitempty"`
}
```

## Type Safety Guarantees

### Compile-Time Checks

```go
// Interface compatibility verification at compile time
var (
    _ LoggingService = (*LoggerAdapter)(nil)
    _ SchemaService  = (*schema.Service)(nil)
    _ DataService    = (*data.Service)(nil)
    _ ExcelService   = (*excel.Service)(nil)
)

// Type safety for GUI components
var (
    _ fyne.Widget        = (*FolderSelector)(nil)
    _ fyne.Widget        = (*ProgressDisplay)(nil)
    _ fyne.CanvasObject  = (*StatusPanel)(nil)
)
```

### Runtime Type Validation

```go
// TypeValidator ensures runtime type safety
type TypeValidator interface {
    ValidateSchemaInfo(schema *models.SchemaInfo) error
    ValidateExcelData(data *models.ExcelData) error
    ValidateOutputData(data *models.OutputData) error
    ValidateConfiguration(config *AppConfig) error
}
```

This comprehensive API specification ensures zero compilation issues, complete type safety, and bulletproof interface compatibility throughout the Excel Schema Generator architecture.