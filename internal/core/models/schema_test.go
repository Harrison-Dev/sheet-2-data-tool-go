package models

import (
	"testing"
	"time"
)

func TestNewSchemaInfo(t *testing.T) {
	schema := NewSchemaInfo()
	
	if schema == nil {
		t.Fatal("NewSchemaInfo returned nil")
	}
	
	if schema.Version == "" {
		t.Error("Schema version should not be empty")
	}
	
	if schema.Files == nil {
		t.Error("Schema files should be initialized")
	}
	
	if len(schema.Files) != 0 {
		t.Error("Schema files should be empty initially")
	}
	
	if schema.CreatedAt.IsZero() {
		t.Error("Schema CreatedAt should be set")
	}
	
	if schema.UpdatedAt.IsZero() {
		t.Error("Schema UpdatedAt should be set")
	}
	
	// Check that timestamps are reasonable (within last second)
	now := time.Now()
	if schema.CreatedAt.After(now) || schema.CreatedAt.Before(now.Add(-time.Second)) {
		t.Errorf("Schema CreatedAt timestamp seems invalid: %v (now: %v)", schema.CreatedAt, now)
	}
}

func TestSchemaInfo_AddFile(t *testing.T) {
	schema := NewSchemaInfo()
	
	fileInfo := ExcelFileInfo{
		FileName: "test.xlsx",
		FilePath: "/path/to/test.xlsx",
		Checksum: "abc123",
		Sheets:   make(map[string]SheetInfo),
	}
	
	schema.AddFile("test.xlsx", fileInfo)
	
	if len(schema.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(schema.Files))
	}
	
	retrieved, exists := schema.Files["test.xlsx"]
	if !exists {
		t.Error("File should exist in schema")
	}
	
	if retrieved.FileName != "test.xlsx" {
		t.Errorf("Expected filename 'test.xlsx', got '%s'", retrieved.FileName)
	}
	
	// UpdatedAt should be updated after CreatedAt
	if !schema.UpdatedAt.After(schema.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt after adding file")
	}
}

func TestSchemaInfo_GetFile(t *testing.T) {
	schema := NewSchemaInfo()
	
	fileInfo := ExcelFileInfo{
		FileName: "test.xlsx",
		FilePath: "/path/to/test.xlsx",
		Checksum: "abc123",
		Sheets:   make(map[string]SheetInfo),
	}
	schema.AddFile("test.xlsx", fileInfo)
	
	retrieved, found := schema.GetFile("test.xlsx")
	
	if !found {
		t.Error("GetFile should return true for existing file")
	}
	
	if retrieved.FileName != "test.xlsx" {
		t.Errorf("Expected filename 'test.xlsx', got '%s'", retrieved.FileName)
	}
	
	if retrieved.Checksum != "abc123" {
		t.Errorf("Expected checksum 'abc123', got '%s'", retrieved.Checksum)
	}
}

func TestSchemaInfo_GetFile_NotFound(t *testing.T) {
	schema := NewSchemaInfo()
	
	retrieved, found := schema.GetFile("nonexistent.xlsx")
	
	if found {
		t.Error("GetFile should return false for non-existent file")
	}
	
	if retrieved.FileName != "" {
		t.Error("Retrieved file should be empty for non-existent file")
	}
}

func TestSchemaInfo_RemoveFile(t *testing.T) {
	schema := NewSchemaInfo()
	
	file1 := ExcelFileInfo{FileName: "test1.xlsx", FilePath: "/path/to/test1.xlsx", Sheets: make(map[string]SheetInfo)}
	file2 := ExcelFileInfo{FileName: "test2.xlsx", FilePath: "/path/to/test2.xlsx", Sheets: make(map[string]SheetInfo)}
	
	schema.AddFile("test1.xlsx", file1)
	schema.AddFile("test2.xlsx", file2)
	
	schema.RemoveFile("test1.xlsx")
	
	if len(schema.Files) != 1 {
		t.Errorf("Expected 1 file after removal, got %d", len(schema.Files))
	}
	
	_, exists := schema.Files["test1.xlsx"]
	if exists {
		t.Error("Removed file should not exist")
	}
	
	_, exists = schema.Files["test2.xlsx"]
	if !exists {
		t.Error("Remaining file should still exist")
	}
}

func TestSchemaInfo_GetSheetCount(t *testing.T) {
	schema := NewSchemaInfo()
	
	sheets1 := make(map[string]SheetInfo)
	sheets1["Sheet1"] = SheetInfo{SheetName: "Sheet1"}
	sheets1["Sheet2"] = SheetInfo{SheetName: "Sheet2"}
	
	sheets2 := make(map[string]SheetInfo)
	sheets2["Data"] = SheetInfo{SheetName: "Data"}
	
	file1 := ExcelFileInfo{
		FileName: "test1.xlsx",
		Sheets:   sheets1,
	}
	
	file2 := ExcelFileInfo{
		FileName: "test2.xlsx",
		Sheets:   sheets2,
	}
	
	schema.AddFile("test1.xlsx", file1)
	schema.AddFile("test2.xlsx", file2)
	
	count := schema.GetSheetCount()
	
	if count != 3 {
		t.Errorf("Expected 3 sheets, got %d", count)
	}
}

func TestSchemaInfo_GetSheetCount_EmptySchema(t *testing.T) {
	schema := NewSchemaInfo()
	
	count := schema.GetSheetCount()
	
	if count != 0 {
		t.Errorf("Expected 0 sheets for empty schema, got %d", count)
	}
}

func TestSchemaInfo_GetFileCount(t *testing.T) {
	schema := NewSchemaInfo()
	
	if schema.GetFileCount() != 0 {
		t.Error("New schema should have 0 files")
	}
	
	file1 := ExcelFileInfo{FileName: "test1.xlsx", Sheets: make(map[string]SheetInfo)}
	file2 := ExcelFileInfo{FileName: "test2.xlsx", Sheets: make(map[string]SheetInfo)}
	
	schema.AddFile("test1.xlsx", file1)
	if schema.GetFileCount() != 1 {
		t.Error("Schema should have 1 file after adding one")
	}
	
	schema.AddFile("test2.xlsx", file2)
	if schema.GetFileCount() != 2 {
		t.Error("Schema should have 2 files after adding two")
	}
}

func TestSchemaInfo_UpdateTimestamp(t *testing.T) {
	schema := NewSchemaInfo()
	
	originalUpdatedAt := schema.UpdatedAt
	
	// Wait a tiny bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)
	
	schema.UpdateTimestamp()
	
	if !schema.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdateTimestamp should update the timestamp")
	}
}

func TestExcelFileInfo_Basic(t *testing.T) {
	sheets := make(map[string]SheetInfo)
	sheets["Sheet1"] = SheetInfo{
		SheetName: "Sheet1",
		ClassName: "TestData",
		DataClass: []DataClassInfo{
			{Name: "ID", DataType: "int", Required: true},
			{Name: "Name", DataType: "string", Required: true},
		},
	}
	
	fileInfo := ExcelFileInfo{
		FileName:    "test.xlsx",
		FilePath:    "/path/to/test.xlsx",
		Checksum:    "abc123",
		Sheets:      sheets,
		LastUpdated: time.Now(),
	}
	
	if fileInfo.FileName != "test.xlsx" {
		t.Errorf("Expected filename 'test.xlsx', got '%s'", fileInfo.FileName)
	}
	
	if len(fileInfo.Sheets) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(fileInfo.Sheets))
	}
	
	sheet, exists := fileInfo.Sheets["Sheet1"]
	if !exists {
		t.Error("Sheet1 should exist")
	}
	
	if len(sheet.DataClass) != 2 {
		t.Errorf("Expected 2 data class fields, got %d", len(sheet.DataClass))
	}
}

func TestDataClassInfo_Basic(t *testing.T) {
	dataClass := DataClassInfo{
		Name:        "TestField",
		DataType:    "string",
		Required:    true,
		Default:     "default_value",
		Description: "Test field description",
	}
	
	if dataClass.Name != "TestField" {
		t.Errorf("Expected name 'TestField', got '%s'", dataClass.Name)
	}
	
	if dataClass.DataType != "string" {
		t.Errorf("Expected data type 'string', got '%s'", dataClass.DataType)
	}
	
	if !dataClass.Required {
		t.Error("Field should be required")
	}
	
	if dataClass.Default != "default_value" {
		t.Errorf("Expected default 'default_value', got '%v'", dataClass.Default)
	}
}

func TestValidationRule_Basic(t *testing.T) {
	rule := ValidationRule{
		Field:      "age",
		Type:       "range",
		Parameters: map[string]interface{}{"min": 18, "max": 65},
	}
	
	if rule.Field != "age" {
		t.Errorf("Expected field 'age', got '%s'", rule.Field)
	}
	
	if rule.Type != "range" {
		t.Errorf("Expected type 'range', got '%s'", rule.Type)
	}
	
	params, ok := rule.Parameters.(map[string]interface{})
	if !ok {
		t.Error("Parameters should be a map")
	}
	
	if params["min"] != 18 {
		t.Errorf("Expected min parameter 18, got %v", params["min"])
	}
}