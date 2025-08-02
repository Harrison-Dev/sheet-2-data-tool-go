package models

import (
	"time"
)

// SchemaInfo represents the complete schema structure for Excel files
type SchemaInfo struct {
	Version   string                    `yaml:"version"`
	Metadata  SchemaMetadata           `yaml:"metadata"`
	Files     map[string]ExcelFileInfo `yaml:"files"`
	CreatedAt time.Time               `yaml:"created_at"`
	UpdatedAt time.Time               `yaml:"updated_at"`
}

// SchemaMetadata contains metadata about the schema
type SchemaMetadata struct {
	Description string   `yaml:"description,omitempty"`
	Author      string   `yaml:"author,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Version     string   `yaml:"schema_version"`
}

// ExcelFileInfo represents information about a single Excel file
type ExcelFileInfo struct {
	FileName    string              `yaml:"file_name"`
	FilePath    string              `yaml:"file_path"`
	Checksum    string              `yaml:"checksum"`
	Sheets      map[string]SheetInfo `yaml:"sheets"`
	LastUpdated time.Time           `yaml:"last_updated"`
}

// SheetInfo represents information about a single Excel sheet
type SheetInfo struct {
	SheetName       string            `yaml:"sheet_name"`
	ClassName       string            `yaml:"class_name"`
	OffsetHeader    int               `yaml:"offset_header"`
	DataClass       []DataClassInfo   `yaml:"data_class"`
	RowCount        int               `yaml:"row_count,omitempty"`
	ValidationRules []ValidationRule  `yaml:"validation_rules,omitempty"`
}

// DataClassInfo represents information about a data field/column
type DataClassInfo struct {
	Name        string      `yaml:"name"`
	DataType    string      `yaml:"data_type"`
	Required    bool        `yaml:"required,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
}

// ValidationRule represents a validation rule for a field
type ValidationRule struct {
	Field      string      `yaml:"field"`
	Type       string      `yaml:"type"`
	Parameters interface{} `yaml:"parameters,omitempty"`
}

// NewSchemaInfo creates a new SchemaInfo with default values
func NewSchemaInfo() *SchemaInfo {
	now := time.Now()
	return &SchemaInfo{
		Version:   "1.0",
		CreatedAt: now,
		UpdatedAt: now,
		Files:     make(map[string]ExcelFileInfo),
		Metadata: SchemaMetadata{
			Version:     "1.0",
			Description: "Generated Excel schema for data conversion",
		},
	}
}

// UpdateTimestamp updates the UpdatedAt field to current time
func (s *SchemaInfo) UpdateTimestamp() {
	s.UpdatedAt = time.Now()
}

// AddFile adds or updates a file in the schema
func (s *SchemaInfo) AddFile(relativePath string, fileInfo ExcelFileInfo) {
	s.Files[relativePath] = fileInfo
	s.UpdateTimestamp()
}

// GetFile retrieves file information by relative path
func (s *SchemaInfo) GetFile(relativePath string) (ExcelFileInfo, bool) {
	fileInfo, exists := s.Files[relativePath]
	return fileInfo, exists
}

// RemoveFile removes a file from the schema
func (s *SchemaInfo) RemoveFile(relativePath string) {
	delete(s.Files, relativePath)
	s.UpdateTimestamp()
}

// GetFileCount returns the number of files in the schema
func (s *SchemaInfo) GetFileCount() int {
	return len(s.Files)
}

// GetSheetCount returns the total number of sheets across all files
func (s *SchemaInfo) GetSheetCount() int {
	count := 0
	for _, fileInfo := range s.Files {
		count += len(fileInfo.Sheets)
	}
	return count
}