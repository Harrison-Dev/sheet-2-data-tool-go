package excelschema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func createTestExcelFile(t *testing.T, filePath string, sheets map[string][][]string) {
	f := excelize.NewFile()
	
	for sheetName, data := range sheets {
		index, err := f.NewSheet(sheetName)
		if err != nil {
			t.Fatalf("Failed to create sheet %s: %v", sheetName, err)
		}
		
		for rowIdx, row := range data {
			for colIdx, value := range row {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
				f.SetCellValue(sheetName, cell, value)
			}
		}
		
		if sheetName != "Sheet1" {
			f.SetActiveSheet(index)
		}
	}
	
	// Remove default Sheet1 if not used
	if _, ok := sheets["Sheet1"]; !ok {
		f.DeleteSheet("Sheet1")
	}
	
	if err := f.SaveAs(filePath); err != nil {
		t.Fatalf("Failed to save Excel file: %v", err)
	}
}

func TestGenerateBasicSchemaFromFolder(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test Excel files
	testFiles := map[string]map[string][][]string{
		"test1.xlsx": {
			"Sheet1": {
				{"ID", "Name", "Value"},
				{"1", "Test", "100"},
				{"2", "Example", "200"},
			},
			"Sheet2": {
				{"Code", "Description"},
				{"A1", "First item"},
				{"B2", "Second item"},
			},
		},
		"subfolder/test2.xlsx": {
			"Data": {
				{"Key", "Data1", "Data2"},
				{"K1", "Value1", "Value2"},
			},
		},
	}
	
	// Create subdirectory
	subDir := filepath.Join(tempDir, "subfolder")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Create Excel files
	for fileName, sheets := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		createTestExcelFile(t, filePath, sheets)
	}
	
	// Create a temporary file that should be ignored
	tempFile := filepath.Join(tempDir, "~$temp.xlsx")
	createTestExcelFile(t, tempFile, map[string][][]string{
		"TempSheet": {{"Should", "Be", "Ignored"}},
	})
	
	// Generate schema
	schema, err := GenerateBasicSchemaFromFolder(tempDir)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	
	// Verify schema structure
	if len(schema.Files) != 2 {
		t.Errorf("Expected 2 files in schema, got %d", len(schema.Files))
	}
	
	// Check test1.xlsx
	if fileInfo, ok := schema.Files["test1.xlsx"]; ok {
		if len(fileInfo.Sheets) != 2 {
			t.Errorf("Expected 2 sheets in test1.xlsx, got %d", len(fileInfo.Sheets))
		}
		
		// Verify Sheet1
		if sheet1, ok := fileInfo.Sheets["Sheet1"]; ok {
			if sheet1.OffsetHeader != 2 {
				t.Errorf("Expected OffsetHeader 2 for Sheet1, got %d", sheet1.OffsetHeader)
			}
			if sheet1.ClassName != "Sheet1" {
				t.Errorf("Expected ClassName 'Sheet1', got '%s'", sheet1.ClassName)
			}
			if sheet1.SheetName != "Sheet1" {
				t.Errorf("Expected SheetName 'Sheet1', got '%s'", sheet1.SheetName)
			}
		} else {
			t.Error("Sheet1 not found in test1.xlsx")
		}
		
		// Verify Sheet2
		if _, ok := fileInfo.Sheets["Sheet2"]; !ok {
			t.Error("Sheet2 not found in test1.xlsx")
		}
	} else {
		t.Error("test1.xlsx not found in schema")
	}
	
	// Check subfolder/test2.xlsx
	if fileInfo, ok := schema.Files[filepath.Join("subfolder", "test2.xlsx")]; ok {
		if len(fileInfo.Sheets) != 1 {
			t.Errorf("Expected 1 sheet in test2.xlsx, got %d", len(fileInfo.Sheets))
		}
		
		if dataSheet, ok := fileInfo.Sheets["Data"]; ok {
			if dataSheet.ClassName != "Data" {
				t.Errorf("Expected ClassName 'Data', got '%s'", dataSheet.ClassName)
			}
		} else {
			t.Error("Data sheet not found in test2.xlsx")
		}
	} else {
		t.Error("subfolder/test2.xlsx not found in schema")
	}
	
	// Verify temporary file was ignored
	for fileName := range schema.Files {
		if filepath.Base(fileName) == "~$temp.xlsx" {
			t.Error("Temporary file was not ignored")
		}
	}
}

func TestGenerateBasicSchemaFromFolder_EmptyFolder(t *testing.T) {
	tempDir := t.TempDir()
	
	schema, err := GenerateBasicSchemaFromFolder(tempDir)
	if err != nil {
		t.Fatalf("Failed to generate schema from empty folder: %v", err)
	}
	
	if len(schema.Files) != 0 {
		t.Errorf("Expected 0 files in schema for empty folder, got %d", len(schema.Files))
	}
}

func TestGenerateBasicSchemaFromFolder_NonExistentFolder(t *testing.T) {
	_, err := GenerateBasicSchemaFromFolder("/non/existent/folder")
	if err == nil {
		t.Error("Expected error for non-existent folder, got nil")
	}
}

func TestProcessExcelFileBasic(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.xlsx")
	
	// Create test Excel file with multiple sheets
	sheets := map[string][][]string{
		"First": {
			{"Col1", "Col2"},
			{"Data1", "Data2"},
		},
		"Second": {
			{"A", "B", "C"},
			{"1", "2", "3"},
		},
		"Third": {
			{"Header"},
			{"Value"},
		},
	}
	
	createTestExcelFile(t, testFile, sheets)
	
	// Process the file
	excelInfo, err := processExcelFileBasic(testFile)
	if err != nil {
		t.Fatalf("Failed to process Excel file: %v", err)
	}
	
	// Verify all sheets are present
	if len(excelInfo.Sheets) != 3 {
		t.Errorf("Expected 3 sheets, got %d", len(excelInfo.Sheets))
	}
	
	// Check each sheet
	expectedSheets := []string{"First", "Second", "Third"}
	for _, sheetName := range expectedSheets {
		sheetInfo, ok := excelInfo.Sheets[sheetName]
		if !ok {
			t.Errorf("Sheet '%s' not found", sheetName)
			continue
		}
		
		if sheetInfo.OffsetHeader != 2 {
			t.Errorf("Expected OffsetHeader 2 for sheet '%s', got %d", sheetName, sheetInfo.OffsetHeader)
		}
		
		if sheetInfo.ClassName != sheetName {
			t.Errorf("Expected ClassName '%s', got '%s'", sheetName, sheetInfo.ClassName)
		}
		
		if sheetInfo.SheetName != sheetName {
			t.Errorf("Expected SheetName '%s', got '%s'", sheetName, sheetInfo.SheetName)
		}
	}
}

func TestProcessExcelFileBasic_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.xlsx")
	
	// Create an invalid file
	err := os.WriteFile(testFile, []byte("This is not a valid Excel file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	_, err = processExcelFileBasic(testFile)
	if err == nil {
		t.Error("Expected error for invalid Excel file, got nil")
	}
}