package excelschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDataFromFolder(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass: []DataClassInfo{
							{Name: "Id", DataType: "int"},
							{Name: "name", DataType: "string"},
							{Name: "price", DataType: "float"},
							{Name: "active", DataType: "bool"},
						},
					},
				},
			},
		},
	}
	
	// Create Excel file
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"Id", "name", "price", "active"},
			{"1", "Product A", "99.99", "true"},
			{"2", "Product B", "149.50", "false"},
			{"3", "Product C", "75.25", "true"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Generate data
	output, err := GenerateDataFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate data: %v", err)
	}
	
	// Verify schema output
	if len(output.Schema) != 1 {
		t.Errorf("Expected 1 schema class, got %d", len(output.Schema))
	}
	
	testDataSchema, ok := output.Schema["TestData"]
	if !ok {
		t.Fatal("TestData schema not found")
	}
	
	if len(testDataSchema) != 4 {
		t.Errorf("Expected 4 fields in schema, got %d", len(testDataSchema))
	}
	
	// Verify field info
	expectedFields := map[string]string{
		"Id":     "int",
		"name":   "string",
		"price":  "float",
		"active": "bool",
	}
	
	for _, field := range testDataSchema {
		expectedType, exists := expectedFields[field.Name]
		if !exists {
			t.Errorf("Unexpected field in schema: %s", field.Name)
			continue
		}
		if field.DataType != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", field.Name, expectedType, field.DataType)
		}
	}
	
	// Verify data output
	if len(output.Data) != 1 {
		t.Errorf("Expected 1 data class, got %d", len(output.Data))
	}
	
	testData, ok := output.Data["TestData"]
	if !ok {
		t.Fatal("TestData not found in output")
	}
	
	if len(testData) != 3 {
		t.Errorf("Expected 3 data rows, got %d", len(testData))
	}
	
	// Verify first row data
	firstRow, ok := testData[0].(map[string]interface{})
	if !ok {
		t.Fatal("First row is not a map")
	}
	
	// Check data types and values
	if id, ok := firstRow["Id"].(int); !ok || id != 1 {
		t.Errorf("Expected Id to be int 1, got %v", firstRow["Id"])
	}
	
	if name, ok := firstRow["name"].(string); !ok || name != "Product A" {
		t.Errorf("Expected name to be string 'Product A', got %v", firstRow["name"])
	}
	
	if price, ok := firstRow["price"].(float64); !ok || price != 99.99 {
		t.Errorf("Expected price to be float64 99.99, got %v", firstRow["price"])
	}
	
	if active, ok := firstRow["active"].(bool); !ok || active != true {
		t.Errorf("Expected active to be bool true, got %v", firstRow["active"])
	}
}

func TestGenerateDataFromFolder_NoIdField(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema without Id field
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Sheet1": {
						OffsetHeader: 1,
						ClassName:    "TestData",
						SheetName:    "Sheet1",
						DataClass: []DataClassInfo{
							{Name: "name", DataType: "string"},
							{Name: "value", DataType: "int"},
						},
					},
				},
			},
		},
	}
	
	// Create Excel file
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Sheet1": {
			{"name", "value"},
			{"Test1", "123"},
			{"Test2", "456"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Generate data - should succeed with auto-generated Id field
	output, err := GenerateDataFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check schema has auto-generated Id field
	schemaFields := output.Schema["TestData"]
	if len(schemaFields) != 3 {
		t.Errorf("Expected 3 fields in schema (including auto-generated Id), got %d", len(schemaFields))
	}
	
	if schemaFields[0].Name != "Id" || schemaFields[0].DataType != "int" {
		t.Errorf("Expected first field to be auto-generated Id field, got: %+v", schemaFields[0])
	}
	
	// Check data has auto-generated Id starting from 0
	data := output.Data["TestData"]
	if len(data) != 2 {
		t.Errorf("Expected 2 rows of data, got %d", len(data))
	}
	
	firstRow := data[0].(map[string]interface{})
	if id, ok := firstRow["Id"].(int); !ok || id != 0 {
		t.Errorf("Expected first row Id to be 0, got %v", firstRow["Id"])
	}
	
	secondRow := data[1].(map[string]interface{})
	if id, ok := secondRow["Id"].(int); !ok || id != 1 {
		t.Errorf("Expected second row Id to be 1, got %v", secondRow["Id"])
	}
}

func TestGenerateDataFromFolder_MultipleSheets(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create schema with multiple sheets
	schema := &SchemaInfo{
		Files: map[string]ExcelFileInfo{
			"test.xlsx": {
				Sheets: map[string]SheetInfo{
					"Users": {
						OffsetHeader: 1,
						ClassName:    "User",
						SheetName:    "Users",
						DataClass: []DataClassInfo{
							{Name: "Id", DataType: "int"},
							{Name: "username", DataType: "string"},
						},
					},
					"Products": {
						OffsetHeader: 1,
						ClassName:    "Product",
						SheetName:    "Products",
						DataClass: []DataClassInfo{
							{Name: "Id", DataType: "int"},
							{Name: "title", DataType: "string"},
							{Name: "price", DataType: "float"},
						},
					},
				},
			},
		},
	}
	
	// Create Excel file with multiple sheets
	excelFile := filepath.Join(tempDir, "test.xlsx")
	sheets := map[string][][]string{
		"Users": {
			{"Id", "username"},
			{"1", "john_doe"},
			{"2", "jane_smith"},
		},
		"Products": {
			{"Id", "title", "price"},
			{"1", "Laptop", "999.99"},
			{"2", "Mouse", "29.99"},
		},
	}
	createTestExcelFile(t, excelFile, sheets)
	
	// Generate data
	output, err := GenerateDataFromFolder(schema, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate data: %v", err)
	}
	
	// Verify both classes are present
	if len(output.Schema) != 2 {
		t.Errorf("Expected 2 schema classes, got %d", len(output.Schema))
	}
	
	if len(output.Data) != 2 {
		t.Errorf("Expected 2 data classes, got %d", len(output.Data))
	}
	
	// Check User data
	userData, ok := output.Data["User"]
	if !ok {
		t.Fatal("User data not found")
	}
	
	if len(userData) != 2 {
		t.Errorf("Expected 2 user records, got %d", len(userData))
	}
	
	// Check Product data
	productData, ok := output.Data["Product"]
	if !ok {
		t.Fatal("Product data not found")
	}
	
	if len(productData) != 2 {
		t.Errorf("Expected 2 product records, got %d", len(productData))
	}
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		value    string
		dataType string
		expected interface{}
		hasError bool
	}{
		{"hello", "string", "hello", false},
		{"123", "int", 123, false},
		{"99.99", "float", 99.99, false},
		{"true", "bool", true, false},
		{"false", "bool", false, false},
		{"invalid", "int", nil, true},
		{"invalid", "float", nil, true},
		{"invalid", "bool", nil, true},
		{"anything", "unknown", "anything", false}, // Unknown types return as string
	}
	
	for _, test := range tests {
		result, err := convertValue(test.value, test.dataType)
		
		if test.hasError && err == nil {
			t.Errorf("Expected error for value '%s' with type '%s', got nil", test.value, test.dataType)
			continue
		}
		
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for value '%s' with type '%s': %v", test.value, test.dataType, err)
			continue
		}
		
		if !test.hasError && result != test.expected {
			t.Errorf("Value '%s' with type '%s': expected %v, got %v", test.value, test.dataType, test.expected, result)
		}
	}
}

func TestSaveJSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.json")
	
	// Create test output
	output := &JSONOutput{
		Schema: map[string][]FieldInfo{
			"TestClass": {
				{Name: "id", DataType: "int"},
				{Name: "name", DataType: "string"},
			},
		},
		Data: map[string][]interface{}{
			"TestClass": {
				map[string]interface{}{
					"id":   1,
					"name": "Test Item",
				},
			},
		},
	}
	
	// Save to file
	err := SaveJSONOutput(output, outputFile)
	if err != nil {
		t.Fatalf("Failed to save JSON output: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}
	
	// Verify content
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	var loadedOutput JSONOutput
	err = json.Unmarshal(data, &loadedOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON output: %v", err)
	}
	
	// Verify schema
	if len(loadedOutput.Schema) != 1 {
		t.Errorf("Expected 1 schema class, got %d", len(loadedOutput.Schema))
	}
	
	testClassSchema, ok := loadedOutput.Schema["TestClass"]
	if !ok {
		t.Fatal("TestClass schema not found")
	}
	
	if len(testClassSchema) != 2 {
		t.Errorf("Expected 2 fields in schema, got %d", len(testClassSchema))
	}
	
	// Verify data
	if len(loadedOutput.Data) != 1 {
		t.Errorf("Expected 1 data class, got %d", len(loadedOutput.Data))
	}
	
	testClassData, ok := loadedOutput.Data["TestClass"]
	if !ok {
		t.Fatal("TestClass data not found")
	}
	
	if len(testClassData) != 1 {
		t.Errorf("Expected 1 data record, got %d", len(testClassData))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 1; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}