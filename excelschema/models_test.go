package excelschema

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestSchemaInfo_SaveToFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_schema.yml")

	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestClass",
						SheetName:    "Sheet1",
						DataClass: []DataClassInfo{
							{Name: "id", DataType: "int"},
							{Name: "name", DataType: "string"},
						},
					},
				},
			},
		},
	}

	err := schema.SaveToFile(testFile)
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Schema file was not created")
	}

	// Verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var loadedSchema SchemaInfo
	err = yaml.Unmarshal(data, &loadedSchema)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved schema: %v", err)
	}

	// Check basic structure
	if len(loadedSchema.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(loadedSchema.Files))
	}

	fileInfo, ok := loadedSchema.Files["test.xlsx"]
	if !ok {
		t.Fatal("test.xlsx not found in loaded schema")
	}

	if len(fileInfo.Sheets) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(fileInfo.Sheets))
	}

	sheetInfo, ok := fileInfo.Sheets["Sheet1"]
	if !ok {
		t.Fatal("Sheet1 not found in loaded schema")
	}

	if sheetInfo.ClassName != "TestClass" {
		t.Errorf("Expected ClassName 'TestClass', got '%s'", sheetInfo.ClassName)
	}

	if len(sheetInfo.DataClass) != 2 {
		t.Errorf("Expected 2 data classes, got %d", len(sheetInfo.DataClass))
	}
}

func TestLoadSchemaFromFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_schema.yml")

	// Create test schema
	originalSchema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 2,
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass: []DataClassInfo{
							{Name: "id", DataType: "int"},
							{Name: "value", DataType: "float"},
							{Name: "active", DataType: "bool"},
						},
					},
				},
			},
		},
	}

	// Save to file
	yamlData, err := yaml.Marshal(originalSchema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	err = os.WriteFile(testFile, yamlData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load from file
	loadedSchema, err := LoadSchemaFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Verify loaded data
	if len(loadedSchema.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(loadedSchema.Files))
	}

	fileInfo, ok := loadedSchema.Files["test.xlsx"]
	if !ok {
		t.Fatal("test.xlsx not found in loaded schema")
	}

	sheetInfo, ok := fileInfo.Sheets["Sheet1"]
	if !ok {
		t.Fatal("Sheet1 not found in loaded schema")
	}

	if sheetInfo.OffsetHeader != 2 {
		t.Errorf("Expected OffsetHeader 2, got %d", sheetInfo.OffsetHeader)
	}

	if len(sheetInfo.DataClass) != 3 {
		t.Errorf("Expected 3 data classes, got %d", len(sheetInfo.DataClass))
	}

	// Verify data types
	expectedTypes := map[string]string{
		"id":     "int",
		"value":  "float",
		"active": "bool",
	}

	for _, dc := range sheetInfo.DataClass {
		expected, ok := expectedTypes[dc.Name]
		if !ok {
			t.Errorf("Unexpected field name: %s", dc.Name)
			continue
		}
		if dc.DataType != expected {
			t.Errorf("Field %s: expected type %s, got %s", dc.Name, expected, dc.DataType)
		}
	}
}

func TestLoadSchemaFromFile_FileNotExist(t *testing.T) {
	_, err := LoadSchemaFromFile("/non/existent/file.yml")
	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}
}

func TestLoadSchemaFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.yml")

	// Write invalid YAML
	err := os.WriteFile(testFile, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadSchemaFromFile(testFile)
	if err == nil {
		t.Error("Expected error when loading invalid YAML, got nil")
	}
}