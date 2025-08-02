package filesystem

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"excel-schema-generator/internal/core/models"
)

func TestNewOutputRepository(t *testing.T) {
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	if repo == nil {
		t.Fatal("NewOutputRepository returned nil")
	}
}

func TestOutputRepository_Save_Success(t *testing.T) {
	// Create test data
	testData := &models.OutputData{
		Version: "1.0",
		Metadata: models.OutputMetadata{
			GeneratedAt: "2024-01-01T00:00:00Z",
			TotalRecords: 2,
		},
		Classes: map[string]*models.DataClass{
			"TestClass": {
				Name: "TestClass",
				Data: []map[string]any{
					{"id": 1, "name": "Test 1"},
					{"id": 2, "name": "Test 2"},
				},
			},
		},
	}
	
	// Create temp directory
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.json")
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testData, outputPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to save output: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
	
	// Verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	
	var savedData models.OutputData
	if err := json.Unmarshal(content, &savedData); err != nil {
		t.Fatalf("Failed to parse saved JSON: %v", err)
	}
	
	if savedData.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", savedData.Version)
	}
	
	if len(savedData.Classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(savedData.Classes))
	}
}

func TestOutputRepository_Save_NilData(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.json")
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, nil, outputPath)
	
	// Assert
	if err == nil {
		t.Error("Expected error for nil data, got nil")
	}
}

func TestOutputRepository_Save_EmptyPath(t *testing.T) {
	testData := &models.OutputData{
		Version: "1.0",
		Classes: map[string]*models.DataClass{},
	}
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testData, "")
	
	// Assert
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
}

func TestOutputRepository_Save_CreateDirectory(t *testing.T) {
	testData := &models.OutputData{
		Version: "1.0",
		Classes: map[string]*models.DataClass{},
	}
	
	// Create nested path that doesn't exist
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "dir", "output.json")
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testData, outputPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to save with directory creation: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created in nested directory")
	}
}

func TestOutputRepository_PrintJSON_Success(t *testing.T) {
	testData := &models.OutputData{
		Version: "1.0",
		Classes: map[string]*models.DataClass{
			"TestClass": {
				Name: "TestClass",
				Data: []map[string]any{
					{"id": 1, "name": "Test"},
				},
			},
		},
	}
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	// PrintJSON writes to stdout, which is hard to test
	// Just verify it doesn't error
	err := repo.PrintJSON(testData)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOutputRepository_PrintJSON_NilData(t *testing.T) {
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	err := repo.PrintJSON(nil)
	
	if err == nil {
		t.Error("Expected error for nil data, got nil")
	}
}

func TestOutputRepository_Save_InvalidPath(t *testing.T) {
	testData := &models.OutputData{
		Version: "1.0",
		Classes: map[string]*models.DataClass{},
	}
	
	// Use an invalid path (directory as file)
	tmpDir := t.TempDir()
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testData, tmpDir) // tmpDir is a directory, not a file
	
	// The behavior depends on the OS
	// Just verify that it handles the error appropriately
	if err == nil {
		// If it succeeds, check if a file was created
		info, err := os.Stat(tmpDir)
		if err == nil && info.IsDir() {
			// It's still a directory, so the save didn't work as expected
			t.Error("Expected error when saving to a directory path")
		}
	}
}

func TestOutputRepository_Save_Overwrite(t *testing.T) {
	// Create existing file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.json")
	
	// Write initial content
	initialContent := []byte(`{"version": "0.9"}`)
	if err := os.WriteFile(outputPath, initialContent, 0644); err != nil {
		t.Fatal(err)
	}
	
	// New data to save
	testData := &models.OutputData{
		Version: "1.0",
		Classes: map[string]*models.DataClass{},
	}
	
	logger := &mockLogger{}
	repo := NewOutputRepository(logger)
	
	ctx := context.Background()
	err := repo.Save(ctx, testData, outputPath)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to overwrite file: %v", err)
	}
	
	// Verify new content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	
	var savedData models.OutputData
	if err := json.Unmarshal(content, &savedData); err != nil {
		t.Fatal(err)
	}
	
	if savedData.Version != "1.0" {
		t.Error("File was not overwritten with new data")
	}
}