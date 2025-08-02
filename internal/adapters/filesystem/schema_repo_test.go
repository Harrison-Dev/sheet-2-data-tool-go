package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"excel-schema-generator/internal/core/models"
	"gopkg.in/yaml.v2"
)

func TestNewSchemaRepository(t *testing.T) {
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	if repo == nil {
		t.Fatal("NewSchemaRepository returned nil")
	}
}

func TestSchemaRepository_Save_Success(t *testing.T) {
	// Create test schema
	testSchema := &models.SchemaInfo{
		Version: "1.0",
		Metadata: models.SchemaMetadata{
			Author:      "Test",
			Description: "Test schema",
		},
		Files: map[string]models.ExcelFileInfo{
			"test.xlsx": {
				FileName: "test.xlsx",
				FilePath: "/path/to/test.xlsx",
				Checksum: "abc123",
				Sheets: map[string]models.SheetInfo{
					"Sheet1": {
						SheetName:    "Sheet1",
						ClassName:    "TestData",
						OffsetHeader: 1,
						DataClass: []models.DataClassInfo{
							{Name: "ID", DataType: "int", Required: true},
							{Name: "Name", DataType: "string", Required: true},
						},
					},
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Create temp directory
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yml")
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testSchema, schemaPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Error("Schema file was not created")
	}
	
	// Verify content
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	
	var savedSchema models.SchemaInfo
	if err := yaml.Unmarshal(content, &savedSchema); err != nil {
		t.Fatalf("Failed to parse saved YAML: %v", err)
	}
	
	if savedSchema.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", savedSchema.Version)
	}
	
	if len(savedSchema.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(savedSchema.Files))
	}
}

func TestSchemaRepository_Save_NilSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yml")
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, nil, schemaPath)
	
	// Assert
	if err == nil {
		t.Error("Expected error for nil schema, got nil")
	}
}

func TestSchemaRepository_Save_EmptyPath(t *testing.T) {
	testSchema := &models.SchemaInfo{
		Version: "1.0",
	}
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testSchema, "")
	
	// Assert
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
}

func TestSchemaRepository_Load_Success(t *testing.T) {
	// Create test schema file
	testSchema := &models.SchemaInfo{
		Version: "1.0",
		Metadata: models.SchemaMetadata{
			Author:      "Test",
			Description: "Test schema",
		},
		Files: map[string]models.ExcelFileInfo{
			"test.xlsx": {
				FileName: "test.xlsx",
				FilePath: "/path/to/test.xlsx",
				Checksum: "abc123",
				Sheets: map[string]models.SheetInfo{
					"Sheet1": {
						SheetName:    "Sheet1",
						ClassName:    "TestData",
						OffsetHeader: 1,
						DataClass: []models.DataClassInfo{
							{Name: "ID", DataType: "int", Required: true},
						},
					},
				},
			},
		},
	}
	
	// Save schema to file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yml")
	
	content, err := yaml.Marshal(testSchema)
	if err != nil {
		t.Fatal(err)
	}
	
	if err := os.WriteFile(schemaPath, content, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Load schema
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	loadedSchema, err := repo.Load(ctx, schemaPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}
	
	if loadedSchema.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", loadedSchema.Version)
	}
	
	if len(loadedSchema.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(loadedSchema.Files))
	}
}

func TestSchemaRepository_Load_FileNotFound(t *testing.T) {
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	_, err := repo.Load(ctx, "/nonexistent/schema.yml")
	
	// Assert
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestSchemaRepository_Load_InvalidYAML(t *testing.T) {
	// Create invalid YAML file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yml")
	
	invalidContent := []byte("invalid: yaml: content: [")
	if err := os.WriteFile(schemaPath, invalidContent, 0644); err != nil {
		t.Fatal(err)
	}
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	_, err := repo.Load(ctx, schemaPath)
	
	// Assert
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestSchemaRepository_Exists_FileExists(t *testing.T) {
	// Create test file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yml")
	
	if err := os.WriteFile(schemaPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	exists, err := repo.Exists(ctx, schemaPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !exists {
		t.Error("Expected file to exist")
	}
}

func TestSchemaRepository_Exists_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "nonexistent.yml")
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	exists, err := repo.Exists(ctx, schemaPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if exists {
		t.Error("Expected file to not exist")
	}
}

func TestSchemaRepository_Save_CreateDirectory(t *testing.T) {
	testSchema := &models.SchemaInfo{
		Version: "1.0",
		Files:   map[string]models.ExcelFileInfo{},
	}
	
	// Create nested path that doesn't exist
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "nested", "dir", "schema.yml")
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testSchema, schemaPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to save with directory creation: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Error("Schema file was not created in nested directory")
	}
}

func TestSchemaRepository_Save_Overwrite(t *testing.T) {
	// Create existing file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yml")
	
	// Write initial content
	initialContent := []byte("version: 0.9")
	if err := os.WriteFile(schemaPath, initialContent, 0644); err != nil {
		t.Fatal(err)
	}
	
	// New schema to save
	testSchema := &models.SchemaInfo{
		Version: "1.0",
		Files:   map[string]models.ExcelFileInfo{},
	}
	
	logger := &mockLogger{}
	repo := NewSchemaRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testSchema, schemaPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to overwrite file: %v", err)
	}
	
	// Verify new content
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	
	var savedSchema models.SchemaInfo
	if err := yaml.Unmarshal(content, &savedSchema); err != nil {
		t.Fatal(err)
	}
	
	if savedSchema.Version != "1.0" {
		t.Error("File was not overwritten with new data")
	}
}