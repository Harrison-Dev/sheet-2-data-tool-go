package excelschema

import (
	"fmt"
	"path/filepath"

	"github.com/sqweek/dialog"
	"github.com/xuri/excelize/v2"
)

func ReadFields(schema *SchemaInfo) error {
	excelDir, err := dialog.Directory().Title("請選擇包含 Excel 文件的資料夾").Browse()
	if err != nil {
		return fmt.Errorf("選擇資料夾時發生錯誤: %v", err)
	}

	if excelDir == "" {
		return fmt.Errorf("沒有選擇資料夾")
	}

	for filePath, fileInfo := range schema.Files {
		fullPath := filepath.Join(excelDir, filePath)
		f, err := excelize.OpenFile(fullPath)
		if err != nil {
			return fmt.Errorf("打開 Excel 文件 %s 時發生錯誤: %v", filePath, err)
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				f.Close()
				return fmt.Errorf("讀取 sheet %s 時發生錯誤: %v", sheetName, err)
			}

			if len(rows) >= sheetInfo.OffsetHeader {
				sheetInfo.DataClass = rows[sheetInfo.OffsetHeader-1]
				fileInfo.Sheets[sheetName] = sheetInfo
			} else {
				fmt.Printf("警告: sheet %s 的行數小於指定的 offset\n", sheetName)
			}
		}

		schema.Files[filePath] = fileInfo
		f.Close()
	}

	return nil
}

func GenerateDataSchema(schema *SchemaInfo) (*SchemaInfo, error) {
	dataSchema := &SchemaInfo{Files: make(map[string]ExcelFileInfo)}

	for filePath, fileInfo := range schema.Files {
		dataFileInfo := ExcelFileInfo{Sheets: make(map[string]SheetInfo)}
		for sheetName, sheetInfo := range fileInfo.Sheets {
			dataSheetInfo := SheetInfo{
				OffsetHeader: sheetInfo.OffsetHeader,
				ClassName:    sheetInfo.ClassName,
				SheetName:    sheetInfo.SheetName,
				DataClass:    sheetInfo.DataClass,
			}
			dataFileInfo.Sheets[sheetName] = dataSheetInfo
		}
		dataSchema.Files[filePath] = dataFileInfo
	}

	return dataSchema, nil
}
