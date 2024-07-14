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
			fmt.Printf("警告: 無法打開 Excel 文件 %s: %v\n", filePath, err)
			continue
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				fmt.Printf("警告: 讀取 sheet %s 時發生錯誤: %v\n", sheetName, err)
				continue
			}

			if len(rows) >= sheetInfo.OffsetHeader {
				headerRow := rows[sheetInfo.OffsetHeader-1]

				excelFields := make(map[string]bool)
				for _, fieldName := range headerRow {
					excelFields[fieldName] = true
				}

				updatedDataClass := []DataClassInfo{}
				for _, dataClass := range sheetInfo.DataClass {
					if excelFields[dataClass.Name] {
						updatedDataClass = append(updatedDataClass, dataClass)
						delete(excelFields, dataClass.Name)
					} else {
						fmt.Printf("信息: 在 sheet %s 中刪除了字段 %s\n", sheetName, dataClass.Name)
					}
				}

				for fieldName := range excelFields {
					updatedDataClass = append(updatedDataClass, DataClassInfo{
						Name:     fieldName,
						DataType: "string",
					})
					fmt.Printf("信息: 在 sheet %s 中新增了字段 %s\n", sheetName, fieldName)
				}

				sheetInfo.DataClass = updatedDataClass

				// 讀取實際數據
				sheetInfo.Data = rows[sheetInfo.OffsetHeader:]

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
				SheetName:    sheetName,
				DataClass:    sheetInfo.DataClass,
				Data:         sheetInfo.Data,
			}
			dataFileInfo.Sheets[sheetName] = dataSheetInfo
		}
		dataSchema.Files[filePath] = dataFileInfo
	}

	return dataSchema, nil
}
