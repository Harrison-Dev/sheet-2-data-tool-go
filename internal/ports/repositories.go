package ports

import (
	"context"
	"io"

	"excel-schema-generator/internal/core/models"
)

// SchemaRepository defines the interface for schema persistence operations
type SchemaRepository interface {
	// Save saves a schema to storage
	Save(ctx context.Context, schema *models.SchemaInfo, path string) error
	
	// Load loads a schema from storage
	Load(ctx context.Context, path string) (*models.SchemaInfo, error)
	
	// Exists checks if a schema exists at the given path
	Exists(ctx context.Context, path string) (bool, error)
	
	// Delete removes a schema from storage
	Delete(ctx context.Context, path string) error
}

// ExcelRepository defines the interface for Excel file operations
type ExcelRepository interface {
	// Read reads an Excel file and returns its data
	Read(ctx context.Context, path string) (*models.ExcelData, error)
	
	// ReadWithOptions reads an Excel file with specific options
	ReadWithOptions(ctx context.Context, path string, options models.ExcelProcessingOptions) (*models.ExcelData, error)
	
	// GetFileInfo retrieves metadata about an Excel file
	GetFileInfo(ctx context.Context, path string) (*models.ExcelFile, error)
	
	// ValidateFile validates that a file is a valid Excel file
	ValidateFile(ctx context.Context, path string) error
}

// FileRepository defines the interface for general file operations
type FileRepository interface {
	// List lists files in a directory with optional pattern matching
	List(ctx context.Context, dir string, pattern string) ([]string, error)
	
	// Exists checks if a file or directory exists
	Exists(ctx context.Context, path string) (bool, error)
	
	// IsDir checks if a path is a directory
	IsDir(ctx context.Context, path string) (bool, error)
	
	// GetInfo retrieves file information
	GetInfo(ctx context.Context, path string) (*FileInfo, error)
	
	// Read reads a file and returns its content
	Read(ctx context.Context, path string) ([]byte, error)
	
	// Write writes content to a file
	Write(ctx context.Context, path string, content []byte) error
	
	// Copy copies a file from source to destination
	Copy(ctx context.Context, src, dst string) error
	
	// Delete removes a file or directory
	Delete(ctx context.Context, path string) error
	
	// CreateDir creates a directory with the given permissions
	CreateDir(ctx context.Context, path string, perm uint32) error
}

// OutputRepository defines the interface for output data persistence
type OutputRepository interface {
	// SaveJSON saves output data as JSON
	SaveJSON(ctx context.Context, output *models.OutputData, path string) error
	
	// SaveWithWriter saves output data using a custom writer
	SaveWithWriter(ctx context.Context, output *models.OutputData, writer io.Writer) error
	
	// LoadJSON loads output data from JSON
	LoadJSON(ctx context.Context, path string) (*models.OutputData, error)
}

// ConfigRepository defines the interface for configuration persistence
type ConfigRepository interface {
	// Save saves configuration to storage
	Save(ctx context.Context, config interface{}, path string) error
	
	// Load loads configuration from storage
	Load(ctx context.Context, path string, config interface{}) error
	
	// Exists checks if configuration exists
	Exists(ctx context.Context, path string) (bool, error)
	
	// GetDefaultPath returns the default configuration path
	GetDefaultPath() string
}

// FileInfo represents basic file information
type FileInfo struct {
	Name         string
	Size         int64
	IsDirectory  bool
	LastModified int64
	Path         string
}

// RepositoryManager defines the interface for managing all repositories
type RepositoryManager interface {
	// Schema returns the schema repository
	Schema() SchemaRepository
	
	// Excel returns the Excel repository
	Excel() ExcelRepository
	
	// File returns the file repository
	File() FileRepository
	
	// Output returns the output repository
	Output() OutputRepository
	
	// Config returns the config repository
	Config() ConfigRepository
	
	// Close closes all repositories and releases resources
	Close() error
}