package excelschema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"excel-schema-generator/pkg/logger"
	"github.com/xuri/excelize/v2"
)

func GenerateBasicSchemaFromFolder(folderPath string) (*SchemaInfo, error) {
	schema := &SchemaInfo{Files: make(map[string]ExcelFileInfo)}

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".xlsx") || strings.HasSuffix(info.Name(), ".xls")) {
			if strings.HasPrefix(info.Name(), "~$") {
				logger.Debug("Skipping temporary file", "file", info.Name())
				return nil
			}

			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				logger.Error("Failed to calculate relative path", "path", path, "folder", folderPath, "error", err)
				return fmt.Errorf("error calculating relative path: %v", err)
			}
			excelInfo, err := processExcelFileBasic(path)
			if err != nil {
				logger.Warn("Error processing file", "file", relativePath, "error", err)
				return nil
			}
			schema.Files[relativePath] = excelInfo
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to scan folder", "folder", folderPath, "error", err)
		return nil, fmt.Errorf("error scanning folder: %v", err)
	}

	return schema, nil
}

func processExcelFileBasic(filePath string) (ExcelFileInfo, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return ExcelFileInfo{}, err
	}
	defer f.Close()

	excelInfo := ExcelFileInfo{Sheets: make(map[string]SheetInfo)}

	for _, sheetName := range f.GetSheetList() {
		excelInfo.Sheets[sheetName] = SheetInfo{
			OffsetHeader: 2,
			ClassName:    sheetName,
			SheetName:    sheetName,
		}
	}

	return excelInfo, nil
}
