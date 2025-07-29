package excelschema

import (
	"path/filepath"
	"testing"
)

func TestUpdateSchemaFromFolder(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create initial schema
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestData",
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
	
	// Create Excel file with updated headers
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"id", "name", "age", "email"}, // Updated headers with new fields
			{"1", "John", "25", "john@example.com"},
			{"2", "Jane", "30", "jane@example.com"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Update schema
	err := UpdateSchemaFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}
	
	// Verify updated schema
	fileInfo := schema.Files["test.xlsx"]
	sheetInfo := fileInfo.Sheets["Sheet1"]
	
	if len(sheetInfo.DataClass) != 4 {
		t.Errorf("Expected 4 data classes after update, got %d", len(sheetInfo.DataClass))
	}
	
	// Check existing fields preserved their types
	expectedFields := map[string]string{
		"id":    "int",    // Should preserve existing type
		"name":  "string", // Should preserve existing type
		"age":   "string", // New field should have default type
		"email": "string", // New field should have default type
	}
	
	for i, dc := range sheetInfo.DataClass {
		expectedType, exists := expectedFields[dc.Name]
		if !exists {
			t.Errorf("Unexpected field '%s' at index %d", dc.Name, i)
			continue
		}
		
		if dc.DataType != expectedType {
			t.Errorf("Field '%s': expected type '%s', got '%s'", dc.Name, expectedType, dc.DataType)
		}
	}
}

func TestUpdateSchemaFromFolder_PreserveExistingFields(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema with custom data types
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass: []DataClassInfo{
							{Name: "id", DataType: "int"},
							{Name: "price", DataType: "float"},
							{Name: "active", DataType: "bool"},
							{Name: "description", DataType: "string"},
						},
					},
				},
			},
		},
	}
	
	// Create Excel file with same headers (should preserve types)
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"id", "price", "active", "description"},
			{"1", "99.99", "true", "Product 1"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Update schema
	err := UpdateSchemaFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}
	
	// Verify all original types are preserved
	sheetInfo := schema.Files["test.xlsx"].Sheets["Sheet1"]
	expectedTypes := map[string]string{
		"id":          "int",
		"price":       "float",
		"active":      "bool",
		"description": "string",
	}
	
	for _, dc := range sheetInfo.DataClass {
		expectedType := expectedTypes[dc.Name]
		if dc.DataType != expectedType {
			t.Errorf("Field '%s': expected type '%s' to be preserved, got '%s'", dc.Name, expectedType, dc.DataType)
		}
	}
}

func TestUpdateSchemaFromFolder_RemovedFields(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema with more fields than Excel file
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass: []DataClassInfo{
							{Name: "id", DataType: "int"},
							{Name: "name", DataType: "string"},
							{Name: "removed_field", DataType: "string"},
							{Name: "another_removed", DataType: "bool"},
						},
					},
				},
			},
		},
	}
	
	// Create Excel file with fewer headers
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"id", "name"}, // Only two fields
			{"1", "Test"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Update schema
	err := UpdateSchemaFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}
	
	// Verify schema only has fields from Excel file
	sheetInfo := schema.Files["test.xlsx"].Sheets["Sheet1"]
	if len(sheetInfo.DataClass) != 2 {
		t.Errorf("Expected 2 data classes after update, got %d", len(sheetInfo.DataClass))
	}
	
	// Check only expected fields remain
	expectedFields := []string{"id", "name"}
	for i, dc := range sheetInfo.DataClass {
		if i < len(expectedFields) {
			if dc.Name != expectedFields[i] {
				t.Errorf("Expected field '%s' at index %d, got '%s'", expectedFields[i], i, dc.Name)
			}
		}
	}
}

func TestUpdateSchemaFromFolder_MultipleSheets(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema with multiple sheets
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "Data1",
						SheetName:    "Sheet1",
						DataClass:    []DataClassInfo{{Name: "id", DataType: "int"}},
					},
					"Sheet2": {
						OffsetHeader: 2,
						ClassName:    "Data2",
						SheetName:    "Sheet2",
						DataClass:    []DataClassInfo{{Name: "code", DataType: "string"}},
					},
				},
			},
		},
	}
	
	// Create Excel file with both sheets
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"id", "value"},
			{"1", "100"},
		},
		"Sheet2": {
			{"header", "Description"},
			{"code", "name"},
			{"A1", "Item A"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Update schema
	err := UpdateSchemaFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}
	
	// Verify Sheet1 updates
	sheet1Info := schema.Files["test.xlsx"].Sheets["Sheet1"]
	if len(sheet1Info.DataClass) != 2 {
		t.Errorf("Sheet1: expected 2 fields, got %d", len(sheet1Info.DataClass))
	}
	
	// Verify Sheet2 updates (note: OffsetHeader is 2, so header is row 2)
	sheet2Info := schema.Files["test.xlsx"].Sheets["Sheet2"]
	if len(sheet2Info.DataClass) != 2 {
		t.Errorf("Sheet2: expected 2 fields, got %d", len(sheet2Info.DataClass))
	}
	
	// Verify original data types preserved
	if sheet1Info.DataClass[0].DataType != "int" {
		t.Errorf("Sheet1 id field type not preserved, got %s", sheet1Info.DataClass[0].DataType)
	}
	
	if sheet2Info.DataClass[0].DataType != "string" {
		t.Errorf("Sheet2 code field type not preserved, got %s", sheet2Info.DataClass[0].DataType)
	}
}

func TestUpdateSchemaFromFolder_InsufficientRows(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema expecting header at row 3
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 3, // Expecting header at row 3
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass:    []DataClassInfo{{Name: "id", DataType: "int"}},
					},
				},
			},
		},
	}
	
	// Create Excel file with only 2 rows
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"First Row"},
			{"Second Row"}, // Only 2 rows, but OffsetHeader expects 3
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Update schema - should handle gracefully
	err := UpdateSchemaFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}
	
	// Schema should remain unchanged since insufficient rows
	sheetInfo := schema.Files["test.xlsx"].Sheets["Sheet1"]
	if len(sheetInfo.DataClass) != 1 {
		t.Errorf("Expected original DataClass to remain unchanged, got %d fields", len(sheetInfo.DataClass))
	}
}

func TestUpdateSchemaFromFolder_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema referencing non-existent file
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"nonexistent.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass:    []DataClassInfo{{Name: "id", DataType: "int"}},
					},
				},
			},
		},
	}
	
	// Update schema - should handle missing file gracefully
	err := UpdateSchemaFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}
	
	// Schema should remain unchanged
	if len(schema.Files) != 1 {
		t.Errorf("Expected 1 file in schema, got %d", len(schema.Files))
	}
}